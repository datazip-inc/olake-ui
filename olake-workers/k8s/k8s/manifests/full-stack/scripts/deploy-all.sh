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
        echo -e "${YELLOW}🔍 Checking logs for $deployment...${NC}"
        kubectl logs deployment/$deployment -n $namespace --tail=20 || true
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
        kubectl logs job/$job -n $namespace || true
        exit 1
    fi
}

# Function to check if PostgreSQL is ready
wait_for_postgres() {
    echo -e "${YELLOW}⏳ Waiting for PostgreSQL to accept connections...${NC}"
    local retries=30
    while [ $retries -gt 0 ]; do
        if kubectl exec deployment/postgresql -n olake -- pg_isready -U temporal -h localhost -p 5432 >/dev/null 2>&1; then
            echo -e "${GREEN}✅ PostgreSQL is ready!${NC}"
            return 0
        fi
        echo -e "${YELLOW}   PostgreSQL not ready yet, waiting... ($retries retries left)${NC}"
        sleep 5
        ((retries--))
    done
    echo -e "${RED}❌ PostgreSQL failed to become ready${NC}"
    exit 1
}

# Function to check if Elasticsearch is ready
wait_for_elasticsearch() {
    echo -e "${YELLOW}⏳ Waiting for Elasticsearch to be ready...${NC}"
    local retries=30
    while [ $retries -gt 0 ]; do
        if kubectl exec deployment/elasticsearch -n olake -- curl -f "http://localhost:9200/_cluster/health?wait_for_status=yellow&timeout=5s" >/dev/null 2>&1; then
            echo -e "${GREEN}✅ Elasticsearch is ready!${NC}"
            return 0
        fi
        echo -e "${YELLOW}   Elasticsearch not ready yet, waiting... ($retries retries left)${NC}"
        sleep 5
        ((retries--))
    done
    echo -e "${RED}❌ Elasticsearch failed to become ready${NC}"
    exit 1
}

# Use minikube's Docker daemon
echo -e "${YELLOW}🐳 Setting up minikube Docker environment...${NC}"
eval $(minikube docker-env)

# Build OLake UI image (main project)
echo -e "${BLUE}🔨 Building OLake UI image (Backend + Frontend)...${NC}"
cd ../../../../  # Go to main project root (olake-ui) - 4 levels up from scripts/
echo -e "${YELLOW}📍 Current directory: $(pwd)${NC}"
echo -e "${YELLOW}🏗️  Building olake-ui:local...${NC}"
docker build -t olake-ui:local .
echo -e "${GREEN}✅ OLake UI image built successfully!${NC}"

# Build K8s worker image
echo -e "${BLUE}🔨 Building OLake K8s Worker image...${NC}"
cd olake-k8s-worker  # Go to olake-k8s-worker directory
echo -e "${YELLOW}📍 Current directory: $(pwd)${NC}"
echo -e "${YELLOW}🏗️  Building olake-k8s-worker:local...${NC}"
docker build -t olake-k8s-worker:local .
echo -e "${GREEN}✅ K8s worker image built successfully!${NC}"

# Go back to manifests directory
cd k8s/manifests/full-stack

# Deploy in order
echo -e "${BLUE}📁 Deploying 00-namespace...${NC}"
kubectl apply -f 00-namespace/

echo -e "${BLUE}🐘 Deploying 01-postgres...${NC}"
kubectl apply -f 01-postgres/secret.yaml
kubectl apply -f 01-postgres/deployment.yaml
kubectl apply -f 01-postgres/service.yaml
wait_for_deployment olake postgresql 300
wait_for_postgres

echo -e "${BLUE}🔍 Deploying 02-elasticsearch...${NC}"
kubectl apply -f 02-elasticsearch/deployment.yaml
kubectl apply -f 02-elasticsearch/service.yaml
wait_for_deployment olake elasticsearch 300
wait_for_elasticsearch

echo -e "${BLUE}⏰ Deploying 03-temporal...${NC}"
kubectl delete deployment temporal temporal-ui -n olake --ignore-not-found=true
kubectl delete job temporal-schema-setup -n olake --ignore-not-found=true

kubectl apply -f 03-temporal/deployment.yaml
kubectl apply -f 03-temporal/ui-deployment.yaml
kubectl apply -f 03-temporal/service.yaml
kubectl apply -f 03-temporal/ui-service.yaml

echo -e "${YELLOW}⏳ Waiting extra time for Temporal auto-setup to initialize database...${NC}"
sleep 30

wait_for_deployment olake temporal 600
wait_for_deployment olake temporal-ui 300

echo -e "${BLUE}🚀 Deploying 04-olake (OLake UI - Backend + Frontend)...${NC}"
# Apply Azure Files PVC first
kubectl apply -f 04-olake/persistent-volume.yaml
kubectl apply -f 04-olake/configmap.yaml
kubectl apply -f 04-olake/deployment.yaml
kubectl apply -f 04-olake/service.yaml
wait_for_deployment olake olake-ui 300

echo -e "${BLUE}👤 Running signup initialization...${NC}"
kubectl delete job olake-signup-init -n olake --ignore-not-found=true
kubectl apply -f 04-olake/init-job.yaml
wait_for_job olake olake-signup-init 120

echo -e "${BLUE}⚡ Deploying 05-olake-worker...${NC}"
kubectl apply -f 05-olake-worker/
wait_for_deployment olake olake-k8s-worker 300

echo -e "${GREEN}🎉 Full stack deployment completed successfully!${NC}"
echo -e "${BLUE}📊 Getting service information...${NC}"

# Show service URLs
echo -e "\n${YELLOW}🔗 Service URLs (using minikube ip):${NC}"
MINIKUBE_IP=$(minikube ip 2>/dev/null || echo "localhost")

echo -e "${GREEN}📱 OLake UI:${NC}         http://$MINIKUBE_IP:30082"
echo -e "${GREEN}🚀 OLake Backend:${NC}    http://$MINIKUBE_IP:30081"
echo -e "${GREEN}⏰ Temporal UI:${NC}     http://$MINIKUBE_IP:30080"

# Show pod status
echo -e "\n${YELLOW}📦 Final Pod Status:${NC}"
kubectl get pods -n olake

echo -e "\n${YELLOW}📋 Useful commands:${NC}"
echo -e "${BLUE}kubectl get pods -n olake${NC}                    # Check all pods"
echo -e "${BLUE}kubectl logs -f deployment/olake-k8s-worker -n olake${NC}  # Worker logs"
echo -e "${BLUE}kubectl logs -f deployment/olake-ui -n olake${NC}           # OLake logs"
echo -e "${BLUE}kubectl logs -f deployment/temporal -n olake${NC}           # Temporal logs"
echo -e "${BLUE}kubectl logs -f deployment/elasticsearch -n olake${NC}      # Elasticsearch logs"
echo -e "${BLUE}kubectl get svc -n olake${NC}                     # Check services"

echo -e "\n${YELLOW}🔧 Troubleshooting:${NC}"
echo -e "${BLUE}kubectl describe pods -n olake${NC}               # Pod details"
echo -e "${BLUE}kubectl get events -n olake --sort-by=.metadata.creationTimestamp${NC}  # Recent events"

echo -e "\n${GREEN}✨ OLake is ready to use!${NC}"
