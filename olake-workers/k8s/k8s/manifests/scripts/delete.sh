#!/bin/bash
# olake-k8s-worker/k8s/delete.sh

echo "Deleting OLake K8s Worker manifests..."

kubectl delete -f manifests/deployment.yaml
kubectl delete -f manifests/service.yaml
kubectl delete -f manifests/configmap.yaml
kubectl delete -f manifests/rbac.yaml
kubectl delete -f manifests/secret.yaml
# Note: namespace left intentionally for other components

echo "OLake K8s Worker deleted successfully!"
