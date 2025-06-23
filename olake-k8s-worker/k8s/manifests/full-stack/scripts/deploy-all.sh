#!/bin/bash
# olake-k8s-worker/k8s/manifests/full-stack/scripts/deploy-all.sh

set -e

echo "🚀 Deploying OLake Full Stack to Minikube..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to wait for deployment
wait_for_deployment() {
    local namespace=$1
    local deployment=$2
    local timeout=${3:-300}
    
    echo -e "${YELLOW}⏳ Waiting for $deployment to be ready...${NC}"
    if kubectl wait --for=condition=available --timeout=${timeout}s deployment/$deployment -n $namespace; then
        echo -e "${GREEN}✅ $deployment is ready!${NC}"
    else
        echo -e "${RED}❌ $deployment failed to become ready within ${timeout}s${NC}"
        exit 1
    fi
}

# Function to wait for job completion
wait_for_job() {
    local namespace=$1
    local job=$2
    local timeout=${3:-300}
    
    echo -e "${YELLOW}⏳ Waiting for $job to complete...${NC}"
    if kubectl wait --for=condition=complete --timeout=${timeout}s job/$job -n $namespace; then
        echo -e "${GREEN}✅ $job completed successfully!${NC}"
    else
        echo -e "${RED}❌ $job failed to complete within ${timeout}s${NC}"
        kubectl logs job/$job -n $namespace
        exit 1
    fi
}

# Deploy in order
echo -e "${BLUE}📁 Deploying 00-namespace...${NC}"
kubectl apply -f 00-namespace/

echo -e "${BLUE}🐘 Deploying 01-postgres...${NC}"
kubectl apply -f 01-postgres/
wait_for_deployment olake postgresql 300

echo -e "${BLUE}⏰ Deploying 02-temporal...${NC}"
kubectl apply -f 02-temporal/
wait_for_deployment olake temporal 600
wait_for_deployment olake temporal-ui 300

echo -e "${BLUE}�� Deploying 03-olake (Backend + Frontend)...${NC}"
kubectl apply -f 03-olake/
wait_for_deployment olake olake 300

echo -e "${BLUE}👤 Running signup initialization...${NC}"
kubectl apply -f 03-olake/init-job.yaml
wait_for_job olake olake-signup-init 120

echo -e "${BLUE}⚡ Deploying 04-olake-worker...${NC}"
kubectl apply -f 04-olake-worker/
wait_for_deployment olake olake-k8s-worker 300

echo -e "${GREEN}🎉 Full stack deployment completed successfully!${NC}"
echo -e "${BLUE}📊 Getting service information...${NC}"

# Show service URLs
echo -e "\n${YELLOW}🔗 Service URLs (using minikube ip):${NC}"
MINIKUBE_IP=$(minikube ip 2>/dev/null || echo "localhost")

echo -e "${GREEN}📱 OLake UI:${NC}         http://$MINIKUBE_IP:30082"
echo -e "${GREEN}🚀 OLake Backend:${NC}    http://$MINIKUBE_IP:30081"
echo -e "${GREEN}⏰ Temporal UI:${NC}     http://$MINIKUBE_IP:30080"

echo -e "\n${YELLOW}📋 Useful commands:${NC}"
echo -e "${BLUE}kubectl get pods -n olake${NC}                    # Check all pods"
echo -e "${BLUE}kubectl logs -f deployment/olake-k8s-worker -n olake${NC}  # Worker logs"
echo -e "${BLUE}kubectl logs -f deployment/olake -n olake${NC}              # OLake logs"
echo -e "${BLUE}kubectl logs -f deployment/temporal -n olake${NC}           # Temporal logs"
echo -e "${BLUE}kubectl get svc -n olake${NC}                     # Check services"

echo -e "\n${GREEN}✨ OLake is ready to use!${NC}"
