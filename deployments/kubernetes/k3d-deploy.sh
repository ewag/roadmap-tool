#!/bin/bash
# K3d deployment script for local testing

set -e

CLUSTER_NAME="${K3D_CLUSTER_NAME:-roadmap-cluster}"
IMAGE_NAME="roadmap-visualizer:latest"

echo "================================================"
echo "K3d Deployment Script for Roadmap Visualizer"
echo "================================================"

# Check if k3d is installed
if ! command -v k3d &> /dev/null; then
    echo "Error: k3d is not installed"
    echo "Install with: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash"
    exit 1
fi

# Check if cluster exists, create if not
if ! k3d cluster list | grep -q "$CLUSTER_NAME"; then
    echo ""
    echo "Creating k3d cluster: $CLUSTER_NAME"
    k3d cluster create $CLUSTER_NAME \
        --agents 1 \
        --port "8080:80@loadbalancer" \
        --wait
    echo "✓ Cluster created"
else
    echo "✓ Using existing cluster: $CLUSTER_NAME"
fi

# Build Docker image
echo ""
echo "Building Docker image..."
cd ../..
docker build -t $IMAGE_NAME .
echo "✓ Image built: $IMAGE_NAME"

# Import image into k3d cluster
echo ""
echo "Importing image into k3d cluster..."
k3d image import $IMAGE_NAME -c $CLUSTER_NAME
echo "✓ Image imported"

# Deploy to Kubernetes
echo ""
echo "Deploying to Kubernetes..."
cd deployments/kubernetes

kubectl apply -f pvc.yaml
echo "✓ PVC created"

kubectl apply -f configmap.yaml
echo "✓ ConfigMap created"

kubectl apply -f deployment.yaml
echo "✓ Deployment created"

kubectl apply -f service.yaml
echo "✓ Service created"

# Wait for deployment
echo ""
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=60s deployment/roadmap-visualizer

# Get service info
echo ""
echo "================================================"
echo "Deployment Complete!"
echo "================================================"
echo ""
echo "Access the application:"
echo "  kubectl port-forward svc/roadmap-visualizer 8080:8080"
echo ""
echo "Then open: http://localhost:8080"
echo ""
echo "Useful commands:"
echo "  kubectl get pods -l app=roadmap-visualizer"
echo "  kubectl logs -l app=roadmap-visualizer -f"
echo "  kubectl exec -it POD_NAME -- /bin/sh"
echo ""
echo "To delete cluster:"
echo "  k3d cluster delete $CLUSTER_NAME"
echo ""
