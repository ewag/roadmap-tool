# Kubernetes Deployment Guide

This guide explains how to deploy the IT Roadmap Visualizer to Kubernetes with persistent storage.

## exports needed

```bash
export GITHUB_USERNAME=<user goes here>
export GITHUB_PAT=<PAT goes here>
```


## Architecture

- **Application**: Go web server serving the roadmap visualizer
- **Storage**: PersistentVolumeClaim (PVC) for storing roadmap data
- **Data Format**: YAML files uploaded via web UI or API
- **Service**: LoadBalancer/NodePort to expose the application

## Prerequisites

- Kubernetes cluster (local or cloud)
- `kubectl` configured to access your cluster
- Docker for building the image
- Container registry access (Docker Hub, GCR, ECR, etc.)

## Quick Start

### 1. Build Docker Image

```bash
# From the project root directory
docker build -t roadmap-visualizer:latest .

# Tag for your registry (replace with your registry)
docker tag roadmap-visualizer:latest YOUR_REGISTRY/roadmap-visualizer:latest

# Push to registry
docker push YOUR_REGISTRY/roadmap-visualizer:latest
```

**For local testing with Minikube/Kind:**
```bash
# Minikube
eval $(minikube docker-env)
docker build -t roadmap-visualizer:latest .

# Kind
docker build -t roadmap-visualizer:latest .
kind load docker-image roadmap-visualizer:latest
```

### 2. Update Deployment Image

Edit `deployments/kubernetes/deployment.yaml` and update the image:

```yaml
containers:
- name: roadmap-visualizer
  image: YOUR_REGISTRY/roadmap-visualizer:latest  # Update this
  imagePullPolicy: Always  # Change to Always for registry images
```

### 3. Deploy to Kubernetes

```bash
cd deployments/kubernetes

# Create PersistentVolumeClaim for data storage
kubectl apply -f pvc.yaml

# Create ConfigMap for configuration
kubectl apply -f configmap.yaml

# Deploy the application
kubectl apply -f deployment.yaml

# Expose the service
kubectl apply -f service.yaml
```

### 4. Access the Application

```bash
# Check pod status
kubectl get pods -l app=roadmap-visualizer

# Check service
kubectl get svc roadmap-visualizer

# For LoadBalancer (cloud environments)
kubectl get svc roadmap-visualizer -o jsonpath='{.status.loadBalancer.ingress[0].ip}'

# For NodePort (local/on-prem)
minikube service roadmap-visualizer --url

# Port forward for testing
kubectl port-forward svc/roadmap-visualizer 8080:8080
```

Then access: http://localhost:8080

## Data Persistence

### How Data is Stored

- All roadmap data is stored in `/data` inside the container
- This directory is mounted from a PVC (default: 1Gi)
- Data persists across pod restarts and redeployments
- Each uploaded YAML creates:
  - A YAML file in `/data/yaml/<uuid>.yaml`
  - A JSON metadata file in `/data/metadata/<uuid>.json`

### Adding Roadmaps

**Via Web UI:**
1. Navigate to http://YOUR_SERVICE_URL/
2. Upload YAML files through the upload form

**Via API:**
```bash
curl -X POST http://YOUR_SERVICE_URL/api/roadmaps \
  -H "Content-Type: application/x-yaml" \
  -H "X-File-Name: my-roadmap.yaml" \
  --data-binary @path/to/roadmap.yaml
```

**Via kubectl exec (direct file placement):**
```bash
# Copy YAML files directly to the PVC
kubectl cp samples/customer-portal.yaml \
  POD_NAME:/data/yaml/customer-portal.yaml

# Then upload via API to create metadata
curl -X POST http://localhost:8080/api/roadmaps \
  -H "Content-Type: application/x-yaml" \
  -H "X-File-Name: customer-portal.yaml" \
  --data-binary @samples/customer-portal.yaml
```

### Backup and Restore

**Backup:**
```bash
# Backup all data from PVC
kubectl exec -it POD_NAME -- tar czf /tmp/backup.tar.gz /data
kubectl cp POD_NAME:/tmp/backup.tar.gz ./roadmap-backup.tar.gz
```

**Restore:**
```bash
# Restore data to PVC
kubectl cp ./roadmap-backup.tar.gz POD_NAME:/tmp/backup.tar.gz
kubectl exec -it POD_NAME -- tar xzf /tmp/backup.tar.gz -C /
kubectl exec -it POD_NAME -- rm /tmp/backup.tar.gz
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `DATA_DIR` | `/data` | Directory for storing roadmap data |

### Resource Limits

Default resources (adjust in `deployment.yaml`):
```yaml
resources:
  requests:
    memory: "64Mi"
    cpu: "100m"
  limits:
    memory: "128Mi"
    cpu: "200m"
```

### Storage Class

Default storage class is `standard`. To use a different class:

```yaml
# pvc.yaml
spec:
  storageClassName: your-storage-class
```

Common options:
- AWS: `gp2`, `gp3`
- GCP: `standard`, `standard-rwo`
- Azure: `default`, `managed-premium`
- Local: `standard`, `local-path`

## Scaling

### Horizontal Scaling

The application supports multiple replicas with shared storage:

```bash
kubectl scale deployment roadmap-visualizer --replicas=3
```

**Note:** With `ReadWriteOnce` PVC, all pods must run on the same node. For multi-node deployments, use:
- `ReadWriteMany` (RWX) storage class (e.g., NFS, EFS, Azure Files)
- Or use ReadWriteOnce with pod affinity to schedule on same node

### Storage Scaling

```bash
# Edit PVC (not all storage classes support expansion)
kubectl edit pvc roadmap-visualizer-pvc

# Change storage request
spec:
  resources:
    requests:
      storage: 5Gi  # Increase from 1Gi
```

## Monitoring

### Health Checks

The application provides:
- `/health` - Liveness probe
- `/ready` - Readiness probe

### Logs

```bash
# View logs
kubectl logs -l app=roadmap-visualizer -f

# View specific pod
kubectl logs POD_NAME -f
```

### Metrics

```bash
# Check resource usage
kubectl top pod -l app=roadmap-visualizer
```

## Troubleshooting

### Pod not starting

```bash
# Check pod status
kubectl describe pod POD_NAME

# Check events
kubectl get events --sort-by='.lastTimestamp'

# Check PVC binding
kubectl get pvc roadmap-visualizer-pvc
```

### PVC issues

```bash
# Check PVC status
kubectl describe pvc roadmap-visualizer-pvc

# Check available storage classes
kubectl get storageclass

# Check PV
kubectl get pv
```

### Data not persisting

```bash
# Verify volume mount
kubectl exec POD_NAME -- ls -la /data

# Check if data directory is writable
kubectl exec POD_NAME -- touch /data/test.txt
kubectl exec POD_NAME -- rm /data/test.txt
```

## Uninstall

```bash
# Delete all resources
kubectl delete -f service.yaml
kubectl delete -f deployment.yaml
kubectl delete -f configmap.yaml

# Delete PVC (this will delete all data!)
kubectl delete -f pvc.yaml
```

## Sample YAML Format

Example roadmap YAML structure:

```yaml
roadmap:
  name: "Project Name"
  service_line: "Team/Department"
  owner: "Owner Name"
  notes: |
    ## Markdown notes
    Support full markdown formatting
  items:
    - id: "unique-id"
      name: "Task Name"
      start: "2025-Q1"
      end: "2025-Q2"
      status: "in-progress"  # planned, in-progress, completed, blocked
      description: "Task description"
      notes: |
        ## Item-level markdown notes
      dependencies: ["other-task-id"]
```

See `samples/` directory for complete examples.

## Advanced Configuration

### Using ConfigMap for Sample Roadmaps

You can pre-populate roadmaps using a ConfigMap (optional):

```bash
kubectl create configmap roadmap-samples --from-file=samples/
```

Then upload them via API on first deployment.

### Ingress

Example ingress configuration:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: roadmap-visualizer
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  rules:
  - host: roadmap.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: roadmap-visualizer
            port:
              number: 8080
  tls:
  - hosts:
    - roadmap.example.com
    secretName: roadmap-tls
```

## Security Considerations

1. **Network Policies**: Restrict pod-to-pod communication
2. **RBAC**: Use service accounts with minimal permissions
3. **PVC Security**: Consider encryption at rest
4. **Ingress**: Use TLS/SSL for external access
5. **Authentication**: Consider adding an auth proxy (OAuth2 Proxy, etc.)

## Support

For issues and questions, refer to the main README.md or open an issue in the repository.
