#!/bin/bash
# olake-k8s-worker/k8s/manifests/full-stack/scripts/cleanup.sh

set -e

echo "🧹 Cleaning up OLake Full Stack from Minikube..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to delete resources safely
safe_delete() {
    local resource_type=$1
    local resource_name=$2
    local namespace=${3:-olake}
    
    if kubectl get $resource_type $resource_name -n $namespace >/dev/null 2>&1; then
        echo -e "${YELLOW}🗑️  Deleting $resource_type/$resource_name...${NC}"
        kubectl delete $resource_type $resource_name -n $namespace --ignore-not-found=true
    fi
}

# Delete in reverse order
echo -e "${BLUE}⚡ Deleting 05-olake-worker...${NC}"
kubectl delete -f 05-olake-worker/ --ignore-not-found=true

echo -e "${BLUE}🌐 Deleting 04-olake-ui...${NC}"
kubectl delete -f 04-olake-ui/ --ignore-not-found=true

echo -e "${BLUE}🚀 Deleting 03-olake-server...${NC}"
kubectl delete -f 03-olake-server/ --ignore-not-found=true

echo -e "${BLUE}⏰ Deleting 02-temporal...${NC}"
kubectl delete -f 02-temporal/ --ignore-not-found=true

echo -e "${BLUE}🐘 Deleting 01-postgres...${NC}"
kubectl delete -f 01-postgres/ --ignore-not-found=true

echo -e "${BLUE}📁 Deleting 00-namespace...${NC}"
kubectl delete -f 00-namespace/ --ignore-not-found=true

# Clean up any orphaned resources
echo -e "${YELLOW}🧽 Cleaning up any orphaned resources...${NC}"

# Wait a bit for graceful deletion
sleep 10

# Force delete namespace if it's stuck
if kubectl get namespace olake >/dev/null 2>&1; then
    echo -e "${YELLOW}🔨 Force deleting namespace...${NC}"
    kubectl delete namespace olake --force --grace-period=0 >/dev/null 2>&1 || true
fi

echo -e "${GREEN}✅ Cleanup completed!${NC}"
echo -e "${BLUE}📊 Remaining resources:${NC}"
kubectl get all -A | grep olake || echo -e "${GREEN}No OLake resources remaining${NC}"
