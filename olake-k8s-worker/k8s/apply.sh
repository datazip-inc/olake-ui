#!/bin/bash
# olake-k8s-worker/k8s/apply.sh

echo "Applying OLake K8s Worker manifests..."

# Apply in correct order
kubectl apply -f manifests/namespace.yaml
kubectl apply -f manifests/secret.yaml
kubectl apply -f manifests/rbac.yaml
kubectl apply -f manifests/configmap.yaml
kubectl apply -f manifests/service.yaml
kubectl apply -f manifests/deployment.yaml

echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/olake-k8s-worker -n olake

echo "OLake K8s Worker deployed successfully!"
