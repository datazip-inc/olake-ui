#!/bin/bash

# Helper script to find Kubernetes service CIDR and suggest a static IP for NFS server
# Usage: ./find-service-ip.sh

set -e
set -o pipefail

echo "🔍 OLake NFS Service IP Helper"
echo "================================="
echo ""

echo "📋 Step 1: Discovering Service CIDR"
echo "-----------------------------------"

# Method 1: Try to infer CIDR from kubernetes service IP (most reliable)
echo "🔍 Checking kubernetes service IP..."
KUBERNETES_IP=$(kubectl get service kubernetes -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")

if [ -n "$KUBERNETES_IP" ]; then
    echo "✅ Found kubernetes service IP: ${KUBERNETES_IP}"
    
    # Infer CIDR from common patterns
    case "$KUBERNETES_IP" in
        10.0.0.1)
            CIDR="10.0.0.0/16"
            ;;
        10.96.0.1)
            CIDR="10.96.0.0/12"
            ;;
        172.20.0.1)
            CIDR="172.20.0.0/16"
            ;;
        100.64.0.1)
            CIDR="100.64.0.0/13"
            ;;
        *)
            # Generic inference: assume /16 for most cases
            IFS='.' read -r o1 o2 o3 o4 <<< "$KUBERNETES_IP"
            if [ "$o4" = "1" ] && [ "$o3" = "0" ]; then
                CIDR="${o1}.${o2}.0.0/16"
            elif [ "$o4" = "1" ]; then
                CIDR="${o1}.${o2}.${o3}.0/24"
            else
                CIDR="${o1}.${o2}.0.0/16"  # Default assumption
            fi
            ;;
    esac
    
    if [ -n "$CIDR" ]; then
        echo "✅ Inferred service CIDR: ${CIDR}"
    fi
fi

# Method 2: API server validation (if inference failed)
if [ -z "$CIDR" ]; then
    echo "🔍 Trying API server validation method..."
    
    # Create a temporary file for the service spec
    TEMP_FILE=$(mktemp)
    cat > "$TEMP_FILE" <<EOF
apiVersion: v1
kind: Service
metadata:
  name: cidr-discovery-dummy-service
spec:
  clusterIP: 1.1.1.1
  ports:
  - port: 80
    protocol: TCP
    targetPort: 80
EOF
    
    # Try to apply with timeout
    CIDR_INFO=$(timeout 10 kubectl apply -f "$TEMP_FILE" 2>&1 || echo "timeout")
    rm -f "$TEMP_FILE"
    
    # Clean up any dummy service
    kubectl delete svc cidr-discovery-dummy-service --ignore-not-found=true >/dev/null 2>&1
    
    if [ "$CIDR_INFO" != "timeout" ]; then
        # Extract CIDR from error message
        CIDR=$(echo "$CIDR_INFO" | grep -o 'The range of valid IPs is [0-9./]*' | sed 's/The range of valid IPs is //' || echo "")
        
        if [ -z "$CIDR" ]; then
            CIDR=$(echo "$CIDR_INFO" | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+' | head -1 || echo "")
        fi
    fi
fi

if [ -z "$CIDR" ]; then
    echo "❌ Could not automatically determine Service CIDR."
    echo ""
    echo "API server response:"
    echo "$CIDR_INFO"
    echo ""
    echo "Please manually find your service CIDR using these cloud-specific commands:"
    echo ""
    echo "🔧 For AKS:"
    echo "   az aks show --resource-group RG --name CLUSTER_NAME --query 'networkProfile.serviceCidr' -o tsv"
    echo ""
    echo "🔧 For EKS:"
    echo "   aws eks describe-cluster --name CLUSTER_NAME --query 'cluster.kubernetesNetworkConfig.serviceCidr' --output text"
    echo ""
    echo "🔧 For GKE:"
    echo "   gcloud container clusters describe CLUSTER_NAME --zone ZONE --format 'value(servicesIpv4CidrBlock)'"
    echo ""
    echo "Then manually choose an IP from that range and use:"
    echo "   helm install olake ./helm/olake --set nfsServer.clusterIP=YOUR_CHOSEN_IP"
    exit 1
fi

echo "✅ Discovered Service CIDR: ${CIDR}"
echo ""

echo "📋 Step 2: Checking Currently Used ClusterIPs"
echo "----------------------------------------------"
echo "Here are the IPs currently in use:"
kubectl get services --all-namespaces -o custom-columns="NAMESPACE:.metadata.namespace,NAME:.metadata.name,CLUSTER-IP:.spec.clusterIP" | grep -v '<none>' | head -10
echo ""

echo "📋 Step 3: Suggesting a Static IP"
echo "----------------------------------"

# Parse CIDR to suggest an IP
IFS='/' read -r IP_BASE MASK <<< "$CIDR"
IFS='.' read -r o1 o2 o3 o4 <<< "$IP_BASE"

# Suggest IP based on common patterns
case "$MASK" in
    12)
        # For /12 networks like 10.96.0.0/12
        SUGGESTED_IP="${o1}.$((o2 + 0)).0.50"
        ;;
    16)
        # For /16 networks like 10.0.0.0/16
        SUGGESTED_IP="${o1}.${o2}.255.50"
        ;;
    *)
        # Default: use a high IP in the range
        SUGGESTED_IP="${o1}.${o2}.${o3}.250"
        ;;
esac

echo "💡 Suggested Static IP: ${SUGGESTED_IP}"
echo ""

echo "📋 Step 4: Verifying the Suggested IP"
echo "--------------------------------------"
echo "Checking if ${SUGGESTED_IP} is available..."

if kubectl get services --all-namespaces | grep -w "${SUGGESTED_IP}" >/dev/null 2>&1; then
    echo "❌ IP ${SUGGESTED_IP} is already in use!"
    echo "Please choose a different IP from the ${CIDR} range."
else
    echo "✅ IP ${SUGGESTED_IP} appears to be available!"
fi

echo ""
echo "📋 Step 5: How to Use This IP"
echo "------------------------------"
echo "Update your values.yaml file:"
echo ""
echo "nfsServer:"
echo "  clusterIP: \"${SUGGESTED_IP}\""
echo ""
echo "Or use it during helm install:"
echo "helm install olake ./olake --set nfsServer.clusterIP=${SUGGESTED_IP}"
echo ""
echo "🔍 To double-check an IP is free:"
echo "kubectl get services --all-namespaces | grep -w \"YOUR_IP\""
echo "(No output means the IP is available)"
echo ""
echo "✅ Done! Use the suggested IP above in your Helm configuration."