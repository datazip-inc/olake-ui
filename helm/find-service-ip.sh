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

# Try the most common method first
CIDR=""
APISERVER_POD=$(kubectl get pods -n kube-system -l component=kube-apiserver -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

if [ -n "$APISERVER_POD" ]; then
    CIDR=$(kubectl get pod "$APISERVER_POD" -n kube-system -o yaml 2>/dev/null | grep 'service-cluster-ip-range' | sed 's/.*=//' | tr -d '"' | head -1 || echo "")
fi

if [ -z "$CIDR" ]; then
    echo "❌ Could not automatically determine Service CIDR from kube-apiserver."
    echo ""
    echo "Please try one of these manual methods:"
    echo ""
    echo "🔧 For GKE:"
    echo "   gcloud container clusters describe CLUSTER_NAME --zone ZONE --format 'value(servicesIpv4CidrBlock)'"
    echo ""
    echo "🔧 For AKS:"
    echo "   az aks show --resource-group RG --name CLUSTER_NAME --query 'networkProfile.serviceCidr' -o tsv"
    echo ""
    echo "🔧 For EKS:"
    echo "   kubectl get service kubernetes -o jsonpath='{.spec.clusterIP}'"
    echo "   (Then infer CIDR from this IP, e.g., 172.20.0.1 → 172.20.0.0/16)"
    echo ""
    echo "🔧 For kubeadm:"
    echo "   kubectl get cm -n kube-system kubeadm-config -o yaml | grep serviceSubnet"
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