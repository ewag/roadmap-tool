#!/bin/bash
# K3d deployment script with GitHub Container Registry support

set -e

CLUSTER_NAME="${K3D_CLUSTER_NAME:-roadmap-cluster}"
GITHUB_USERNAME="${GITHUB_USERNAME:-}"
GITHUB_PAT="${GITHUB_PAT:-}"
IMAGE_TAG="${IMAGE_TAG:-latest}"

echo "================================================"
echo "K3d + GitHub Container Registry Deployment"
echo "================================================"

# Check if k3d is installed
if ! command -v k3d &> /dev/null; then
    echo "Error: k3d is not installed"
    echo "Install with: curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash"
    exit 1
fi

# Check for required environment variables
if [ -z "$GITHUB_USERNAME" ] || [ -z "$GITHUB_PAT" ]; then
    echo "Error: GITHUB_USERNAME and GITHUB_PAT environment variables are required"
    echo ""
    echo "Usage:"
    echo "  export GITHUB_USERNAME=your-username"
    echo "  export GITHUB_PAT=your-personal-access-token"
    echo "  ./k3d-ghcr-deploy.sh"
    echo ""
    echo "To create a GitHub PAT:"
    echo "  1. Go to GitHub Settings > Developer settings > Personal access tokens"
    echo "  2. Generate new token (classic)"
    echo "  3. Select scope: read:packages"
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

# Set kubectl context
kubectl config use-context k3d-$CLUSTER_NAME

# Create namespace if it doesn't exist
kubectl create namespace roadmap-visualizer --dry-run=client -o yaml | kubectl apply -f -

# Create image pull secret
echo ""
echo "Creating image pull secret..."
kubectl delete secret ghcr-secret -n roadmap-visualizer --ignore-not-found
kubectl create secret docker-registry ghcr-secret \
    --docker-server=ghcr.io \
    --docker-username="$GITHUB_USERNAME" \
    --docker-password="$GITHUB_PAT" \
    --docker-email="${GITHUB_EMAIL:-$GITHUB_USERNAME@users.noreply.github.com}" \
    --namespace=roadmap-visualizer
echo "✓ Image pull secret created"

# Update deployment with correct image
IMAGE_NAME="ghcr.io/$GITHUB_USERNAME/roadmap_tool:$IMAGE_TAG"
echo ""
echo "Using image: $IMAGE_NAME"

cd deployments/kubernetes

# Apply Kubernetes manifests
echo ""
echo "Deploying to Kubernetes..."

kubectl apply -f pvc.yaml -n roadmap-visualizer
echo "✓ PVC created"

kubectl apply -f configmap.yaml -n roadmap-visualizer
echo "✓ ConfigMap created"

# Update image in deployment
cat deployment.yaml | \
    sed "s|image: .*|image: $IMAGE_NAME|" | \
    sed "s|imagePullPolicy: .*|imagePullPolicy: Always|" | \
    kubectl apply -f - -n roadmap-visualizer
echo "✓ Deployment created"

kubectl apply -f service.yaml -n roadmap-visualizer
echo "✓ Service created"

# Wait for deployment
echo ""
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=120s deployment/roadmap-visualizer -n roadmap-visualizer

# Get pods
echo ""
echo "Pods:"
kubectl get pods -n roadmap-visualizer -l app=roadmap-visualizer

# Get service info
echo ""
echo "================================================"
echo "Deployment Complete!"
echo "================================================"
echo ""
echo "Access the application:"
echo "  kubectl port-forward -n roadmap-visualizer svc/roadmap-visualizer 8080:8080"
echo ""
echo "Then open: http://localhost:8080"
echo ""
echo "Useful commands:"
echo "  kubectl get pods -n roadmap-visualizer -l app=roadmap-visualizer"
echo "  kubectl logs -n roadmap-visualizer -l app=roadmap-visualizer -f"
echo "  kubectl describe pod -n roadmap-visualizer POD_NAME"
echo ""
echo "To delete:"
echo "  k3d cluster delete $CLUSTER_NAME"
echo ""
