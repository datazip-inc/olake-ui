package activities

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"olake-k8s-worker/logger"
	"olake-k8s-worker/shared"
	"olake-k8s-worker/utils"
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
	namespace := utils.GetEnv("WORKER_NAMESPACE", "default")

	logger.Infof("Initialized K8s job manager for namespace: %s", namespace)

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

	logger.Debugf("Creating ConfigMap %s in namespace %s", name, k.namespace)
	result, err := k.clientset.CoreV1().ConfigMaps(k.namespace).Create(ctx, configMap, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("Failed to create ConfigMap %s: %v", name, err)
		return nil, err
	}

	logger.Infof("Successfully created ConfigMap %s", name)
	return result, nil
}

// CreateJob creates a Kubernetes Job for running sync operations
func (k *K8sJobManager) CreateJob(ctx context.Context, spec *JobSpec) (*batchv1.Job, error) {
	// Get TTL from environment
	ttlSeconds := utils.GetEnvInt("JOB_TTL_SECONDS", 0)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":       "olake-sync",
				"type":      "connector-job",
				"cleanup":   "auto",
				"operation": string(spec.Operation),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &[]int32{1}[0],
			// Only set TTL if > 0
			TTLSecondsAfterFinished: getTTLPointer(ttlSeconds),
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
									corev1.ResourceMemory: utils.ParseQuantity("256Mi"),
									corev1.ResourceCPU:    utils.ParseQuantity("100m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: utils.ParseQuantity("1Gi"),
									corev1.ResourceCPU:    utils.ParseQuantity("500m"),
								},
							},
						},
					},
				},
			},
		},
	}

	logger.Infof("Creating Job %s with image %s", spec.Name, spec.Image)
	result, err := k.clientset.BatchV1().Jobs(k.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("Failed to create Job %s: %v", spec.Name, err)
		return nil, err
	}

	logger.Infof("Successfully created Job %s", spec.Name)
	return result, nil
}

// WaitForJobCompletion waits for a Job to complete and returns the result
func (k *K8sJobManager) WaitForJobCompletion(ctx context.Context, jobName string, timeout time.Duration) (map[string]interface{}, error) {
	logger.Infof("Waiting for Job %s to complete (timeout: %v)", jobName, timeout)
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		job, err := k.clientset.BatchV1().Jobs(k.namespace).Get(ctx, jobName, metav1.GetOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to get job status: %v", err)
		}

		// Check if job completed successfully
		if job.Status.Succeeded > 0 {
			logger.Infof("Job %s completed successfully", jobName)
			// Get pod logs to extract results
			return k.getJobResults(ctx, jobName)
		}

		// Check if job failed
		if job.Status.Failed > 0 {
			logger.Errorf("Job %s failed", jobName)
			logs, _ := k.getJobLogs(ctx, jobName)
			return nil, fmt.Errorf("job failed: %s", logs)
		}

		// Wait before checking again
		time.Sleep(5 * time.Second)
	}

	logger.Errorf("Job %s timed out after %v", jobName, timeout)
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
	logger.Debugf("Getting results from pod %s for job %s", podName, jobName)

	// Get pod logs
	logs, err := k.getPodLogs(ctx, podName)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %v", err)
	}

	logger.Debugf("Raw pod logs for job %s:\n%s", jobName, logs)

	// Use flexible parsing that can handle different output formats
	return utils.ParseJobOutput(logs)
}

// CleanupJob removes a job and its associated ConfigMap
func (k *K8sJobManager) CleanupJob(ctx context.Context, jobName, configMapName string) error {
	logger.Infof("Cleaning up Job %s and ConfigMap %s", jobName, configMapName)

	// Delete job
	propagationPolicy := metav1.DeletePropagationForeground
	err := k.clientset.BatchV1().Jobs(k.namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		logger.Errorf("Failed to delete job %s: %v", jobName, err)
		return fmt.Errorf("failed to delete job: %v", err)
	}

	// Delete ConfigMap
	err = k.clientset.CoreV1().ConfigMaps(k.namespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		logger.Errorf("Failed to delete ConfigMap %s: %v", configMapName, err)
		return fmt.Errorf("failed to delete configmap: %v", err)
	}

	logger.Infof("Successfully cleaned up Job %s and ConfigMap %s", jobName, configMapName)
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
	Name      string
	Image     string
	Command   []string
	Args      []string
	Operation shared.Command
}

// Helper function to return TTL pointer only if > 0
func getTTLPointer(ttlSeconds int) *int32 {
	if ttlSeconds <= 0 {
		return nil // No TTL - job persists indefinitely
	}
	ttl := int32(ttlSeconds)
	return &ttl
}

// CreateJobWithPV creates a Kubernetes Job for running sync operations with PV
func (k *K8sJobManager) CreateJobWithPV(ctx context.Context, spec *JobSpec, configs []shared.JobConfig) (*batchv1.Job, error) {
	// Create unique directory for this workflow
	workflowDir := spec.Name // e.g., "sync-mysql-1234567890"

	// Write config files to PV (similar to Docker implementation)
	if err := k.setupWorkDirectory(workflowDir); err != nil {
		return nil, fmt.Errorf("failed to setup work directory: %v", err)
	}

	if err := k.writeConfigFiles(workflowDir, configs); err != nil {
		return nil, fmt.Errorf("failed to write config files: %v", err)
	}

	// Create Job with PV mount instead of ConfigMap
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":       "olake-sync",
				"type":      "connector-job",
				"cleanup":   "auto",
				"operation": string(spec.Operation),
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &[]int32{1}[0],
			// Only set TTL if > 0
			TTLSecondsAfterFinished: getTTLPointer(utils.GetEnvInt("JOB_TTL_SECONDS", 0)),
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
									Name:      "job-storage",
									MountPath: "/mnt/config",
									SubPath:   workflowDir, // Mount specific workflow directory
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: utils.ParseQuantity("256Mi"),
									corev1.ResourceCPU:    utils.ParseQuantity("100m"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: utils.ParseQuantity("1Gi"),
									corev1.ResourceCPU:    utils.ParseQuantity("500m"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "job-storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "olake-jobs-pvc",
								},
							},
						},
					},
				},
			},
		},
	}

	logger.Infof("Creating Job %s with image %s", spec.Name, spec.Image)
	result, err := k.clientset.BatchV1().Jobs(k.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		logger.Errorf("Failed to create Job %s: %v", spec.Name, err)
		return nil, err
	}

	logger.Infof("Successfully created Job %s", spec.Name)
	return result, nil
}

func (k *K8sJobManager) setupWorkDirectory(workflowDir string) error {
	// Use similar logic to Docker runner
	basePath := "/data/olake-jobs" // PV mount point on worker pod
	workDir := filepath.Join(basePath, workflowDir)

	return utils.CreateDirectory(workDir, 0755)
}

func (k *K8sJobManager) writeConfigFiles(workflowDir string, configs []shared.JobConfig) error {
	basePath := "/data/olake-jobs"
	workDir := filepath.Join(basePath, workflowDir)

	for _, config := range configs {
		filePath := filepath.Join(workDir, config.Name)
		if err := utils.WriteFile(filePath, []byte(config.Data), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", config.Name, err)
		}
	}
	return nil
}
