package activities

import (
	"context"
	"fmt"
	"os"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"olake-k8s-worker/shared"
)

// K8sJobManager handles Kubernetes Job operations
type K8sJobManager struct {
	clientset kubernetes.Interface
	namespace string
}

// NewK8sJobManager creates a new Kubernetes Job manager
func NewK8sJobManager() (*K8sJobManager, error) {
	// Use in-cluster configuration
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	// Get namespace from environment or use default
	namespace := "default"
	if ns := GetEnv("WORKER_NAMESPACE", ""); ns != "" {
		namespace = ns
	}

	return &K8sJobManager{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

// GetDockerImageName constructs a Docker image name based on source type and version
func (k *K8sJobManager) GetDockerImageName(sourceType, version string) string {
	if version == "" {
		version = "latest"
	}
	return fmt.Sprintf("olakego/source-%s:%s", sourceType, version)
}

// CreateConfigMap creates a ConfigMap with job configuration files
func (k *K8sJobManager) CreateConfigMap(ctx context.Context, name string, configs []shared.JobConfig) (*corev1.ConfigMap, error) {
	data := make(map[string]string)
	for _, config := range configs {
		data[config.Name] = config.Data
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":     "olake-sync",
				"type":    "job-config",
				"cleanup": "auto",
			},
		},
		Data: data,
	}

	return k.clientset.CoreV1().ConfigMaps(k.namespace).Create(ctx, configMap, metav1.CreateOptions{})
}

// CreateJob creates a Kubernetes Job for running sync operations
func (k *K8sJobManager) CreateJob(ctx context.Context, spec *JobSpec) (*batchv1.Job, error) {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":       "olake-sync",
				"type":      "connector-job",
				"cleanup":   "auto",
				"operation": string(spec.Command),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &[]int32{1}[0], // Match Temporal retry policy
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "connector",
							Image:   spec.Image,
							Command: spec.Command,
							Args:    spec.Args,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "config",
									MountPath: "/mnt/config",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: ParseQuantity("256Mi"),
									corev1.ResourceCPU:    ParseQuantity("100m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: ParseQuantity("1Gi"),
									corev1.ResourceCPU:    ParseQuantity("500m"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: spec.ConfigMapName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return k.clientset.BatchV1().Jobs(k.namespace).Create(ctx, job, metav1.CreateOptions{})
}

// WaitForJobCompletion waits for a Job to complete and returns the result
func (k *K8sJobManager) WaitForJobCompletion(ctx context.Context, jobName string, timeout time.Duration) (map[string]interface{}, error) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		job, err := k.clientset.BatchV1().Jobs(k.namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get job status: %v", err)
		}

		// Check if job completed successfully
		if job.Status.Succeeded > 0 {
			// Get pod logs to extract results
			return k.getJobResults(ctx, jobName)
		}

		// Check if job failed
		if job.Status.Failed > 0 {
			logs, _ := k.getJobLogs(ctx, jobName)
			return nil, fmt.Errorf("job failed: %s", logs)
		}

		// Wait before checking again
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("job timed out after %v", timeout)
}

// getJobResults extracts results from completed job
func (k *K8sJobManager) getJobResults(ctx context.Context, jobName string) (map[string]interface{}, error) {
	// Get the pod associated with the job
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods for job: %v", err)
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("no pods found for job")
	}

	podName := pods.Items[0].Name

	// Get pod logs
	logs, err := k.getPodLogs(ctx, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %v", err)
	}

	// Parse the logs to extract results (similar to Docker implementation)
	return ParseJobOutput(logs)
}

// CleanupJob removes a job and its associated ConfigMap
func (k *K8sJobManager) CleanupJob(ctx context.Context, jobName, configMapName string) error {
	// Delete job
	propagationPolicy := metav1.DeletePropagationForeground
	err := k.clientset.BatchV1().Jobs(k.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job: %v", err)
	}

	// Delete ConfigMap
	err = k.clientset.CoreV1().ConfigMaps(k.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete configmap: %v", err)
	}

	return nil
}

// Helper functions

func (k *K8sJobManager) getPodLogs(ctx context.Context, podName string) (string, error) {
	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(podName, &corev1.PodLogOptions{})
	logs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer logs.Close()

	buf := make([]byte, 4096)
	var result string
	for {
		n, err := logs.Read(buf)
		if n > 0 {
			result += string(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return result, nil
}

func (k *K8sJobManager) getJobLogs(ctx context.Context, jobName string) (string, error) {
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("job-name=%s", jobName),
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no pods found for job")
	}

	return k.getPodLogs(ctx, pods.Items[0].Name)
}

// JobSpec defines the specification for creating a Kubernetes Job
type JobSpec struct {
	Name          string
	Image         string
	Command       []string
	Args          []string
	ConfigMapName string
	Command       shared.Command
}

// Utility functions
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func ParseQuantity(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q
}
