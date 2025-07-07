#!/bin/bash
# olake-k8s-worker/k8s/manifests/full-stack/scripts/status.sh

set -e

echo "📊 OLake Full Stack Status"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if namespace exists
if ! kubectl get namespace olake >/dev/null 2>&1; then
    echo -e "${RED}❌ OLake namespace not found. Run deploy-all.sh first.${NC}"
    exit 1
fi

echo -e "\n${BLUE}📦 Pods Status:${NC}"
kubectl get pods -n olake -o wide

echo -e "\n${BLUE}🔧 Services:${NC}"
kubectl get svc -n olake

echo -e "\n${BLUE}⚙️  Deployments:${NC}"
kubectl get deployments -n olake

echo -e "\n${BLUE}📊 ConfigMaps:${NC}"
kubectl get configmaps -n olake

echo -e "\n${BLUE}🔐 Secrets:${NC}"
kubectl get secrets -n olake

# Check service URLs
echo -e "\n${YELLOW}🔗 Service URLs:${NC}"
MINIKUBE_IP=$(minikube ip 2>/dev/null || echo "localhost")

echo -e "${GREEN}📱 OLake UI:${NC}         http://$MINIKUBE_IP:30082"
echo -e "${GREEN}🚀 OLake Backend:${NC}    http://$MINIKUBE_IP:30081"
echo -e "${GREEN}⏰ Temporal UI:${NC}     http://$MINIKUBE_IP:30080"

# Check individual service health
echo -e "\n${YELLOW}🏥 Health Checks:${NC}"

# Check if services are responding
services=("30082:OLake UI" "30081:OLake Backend" "30080:Temporal UI")

for service in "${services[@]}"; do
    port=$(echo $service | cut -d: -f1)
    name=$(echo $service | cut -d: -f2)
    
    if curl -s --max-time 5 http://$MINIKUBE_IP:$port >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $name is responding${NC}"
    else
        echo -e "${RED}❌ $name is not responding${NC}"
    fi
done
