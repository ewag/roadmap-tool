# GitHub Actions + GHCR + k3d Setup Guide

This guide explains how to set up automated Docker builds with GitHub Actions, push to GitHub Container Registry (GHCR), and deploy to k3d.

## Overview

The workflow:
1. **GitHub Actions** builds Docker image on push to main
2. **GHCR** stores the container image
3. **k3d** pulls and deploys the image using an image pull secret

## Step 1: GitHub Actions Setup

### Enable GitHub Container Registry

1. Go to your repository on GitHub
2. Settings → Actions → General
3. Under "Workflow permissions", select:
   - ✅ "Read and write permissions"
   - ✅ "Allow GitHub Actions to create and approve pull requests"

### Push the Workflow

The workflow is already created at `.github/workflows/docker-build.yml`. Just commit and push:

```bash
git add .github/workflows/docker-build.yml
git commit -m "Add GitHub Actions Docker build workflow"
git push origin main
```

The workflow will:
- Build on push to `main` or `develop` branches
- Build on tags starting with `v*` (e.g., `v1.0.0`)
- Create multi-platform images (amd64 + arm64)
- Push to `ghcr.io/YOUR_USERNAME/roadmap_tool`

### Workflow Triggers

- **On push to main**: Builds and tags as `latest` + commit SHA
- **On tags**: Builds semantic versions (`v1.0.0` → `1.0.0`, `1.0`, `1`)
- **On PR**: Builds but doesn't push (validation only)
- **Manual**: Can trigger via Actions tab

## Step 2: Create GitHub Personal Access Token (PAT)

To pull images from GHCR in k3d, you need a PAT:

1. Go to GitHub Settings (your user, not repository)
2. Developer settings → Personal access tokens → Tokens (classic)
3. Generate new token (classic)
4. Select scopes:
   - ✅ `read:packages` - Download packages from GitHub Package Registry
   - ✅ `write:packages` - Upload packages to GitHub Package Registry (optional)
5. Generate and **copy the token** (you won't see it again!)

## Step 3: Deploy to k3d

### Option A: Automated Script (Recommended)

```bash
# Set environment variables
export GITHUB_USERNAME=your-username
export GITHUB_PAT=ghp_your_token_here

# Run deployment script
cd deployments/kubernetes
./k3d-ghcr-deploy.sh
```

This script will:
- Create k3d cluster (if needed)
- Create image pull secret
- Deploy all Kubernetes manifests
- Pull image from GHCR
- Wait for deployment to be ready

### Option B: Manual Deployment

#### 1. Create k3d Cluster

```bash
k3d cluster create roadmap-cluster \
  --agents 1 \
  --port "8080:80@loadbalancer" \
  --wait
```

#### 2. Create Image Pull Secret

```bash
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_PAT \
  --docker-email=your-email@example.com
```

#### 3. Update Deployment YAML

Edit `deployment.yaml`:

```yaml
spec:
  imagePullSecrets:
  - name: ghcr-secret
  containers:
  - name: roadmap-visualizer
    image: ghcr.io/YOUR_GITHUB_USERNAME/roadmap_tool:latest
    imagePullPolicy: Always
```

#### 4. Deploy

```bash
cd deployments/kubernetes
kubectl apply -f pvc.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

#### 5. Access Application

```bash
kubectl port-forward svc/roadmap-visualizer 8080:8080
```

Open: http://localhost:8080

## Step 4: Verify Image in GHCR

After the GitHub Action completes:

1. Go to your GitHub profile
2. Packages tab
3. You should see `roadmap_tool` package
4. Click it → Package settings → Change visibility to Public (optional)

**View available tags:**
```bash
# List all tags
curl -H "Authorization: token YOUR_GITHUB_PAT" \
  https://api.github.com/users/YOUR_USERNAME/packages/container/roadmap_tool/versions
```

## Deployment Strategies

### Using Latest Tag (Development)

```yaml
image: ghcr.io/YOUR_USERNAME/roadmap_tool:latest
```

Always pulls the most recent build from main branch.

### Using Version Tags (Production)

```bash
# Create a version tag
git tag v1.0.0
git push origin v1.0.0
```

Then deploy:
```yaml
image: ghcr.io/YOUR_USERNAME/roadmap_tool:1.0.0
```

### Using Commit SHA (Immutable)

```yaml
image: ghcr.io/YOUR_USERNAME/roadmap_tool:main-abc1234
```

Most reliable for production - image never changes.

## Troubleshooting

### Image Pull Errors

**Error:** `ErrImagePull` or `ImagePullBackOff`

Check:
```bash
# Verify secret exists
kubectl get secret ghcr-secret

# Describe pod for error details
kubectl describe pod POD_NAME

# Check secret format
kubectl get secret ghcr-secret -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d
```

**Solutions:**
1. Verify PAT has `read:packages` scope
2. Check if package is private and PAT has access
3. Verify username/token are correct
4. Recreate the secret

### GitHub Actions Not Building

Check:
1. Actions tab → View workflow runs
2. Workflow file syntax (YAML format)
3. Repository permissions (Settings → Actions)
4. Check if workflow file is in `.github/workflows/`

### Image Not Found in GHCR

1. Check Actions tab - did build complete successfully?
2. Verify package visibility (Profile → Packages)
3. Check image name matches repository name
4. For private repos, package is private by default

### Update Image in Running Cluster

After pushing new code:

```bash
# Wait for GitHub Action to complete
# Then restart deployment
kubectl rollout restart deployment/roadmap-visualizer

# Or force pull latest
kubectl delete pod -l app=roadmap-visualizer
```

## Best Practices

### 1. Use Semantic Versioning

```bash
git tag -a v1.2.3 -m "Release version 1.2.3"
git push origin v1.2.3
```

Benefits:
- Immutable deployments
- Easy rollback
- Clear version tracking

### 2. Separate Environments

```yaml
# Development
image: ghcr.io/user/roadmap_tool:latest

# Staging
image: ghcr.io/user/roadmap_tool:develop

# Production
image: ghcr.io/user/roadmap_tool:1.2.3
```

### 3. Image Cleanup

GHCR has storage limits. Clean old images:

1. Go to Package settings
2. Select old versions
3. Delete unused tags

Or use GitHub's cleanup action: `ghcr.io/actions/delete-package-versions`

### 4. Security

- Don't commit PAT tokens
- Use GitHub Secrets for CI/CD
- Regularly rotate PATs
- Use minimal scopes (`read:packages` only for deployments)
- Enable 2FA on GitHub account

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GITHUB_USERNAME` | Your GitHub username | `john-doe` |
| `GITHUB_PAT` | Personal access token | `ghp_abc123...` |
| `IMAGE_TAG` | Image tag to deploy | `latest`, `1.0.0` |
| `K3D_CLUSTER_NAME` | k3d cluster name | `roadmap-cluster` |

## Monitoring GitHub Actions

### View Build Logs

1. GitHub → Actions tab
2. Click on workflow run
3. Click on job name
4. Expand steps to see logs

### GitHub Actions Status Badge

Add to README.md:

```markdown
![Build Status](https://github.com/YOUR_USERNAME/roadmap_tool/workflows/Build%20and%20Push%20Docker%20Image/badge.svg)
```

## Cost Considerations

- **GitHub Actions**: 2,000 free minutes/month for private repos (unlimited for public)
- **GHCR Storage**: 500MB free, then $0.25/GB
- **Bandwidth**: Unlimited for public packages

## Additional Resources

- [GitHub Container Registry docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [GitHub Actions docs](https://docs.github.com/en/actions)
- [k3d documentation](https://k3d.io/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)

## Quick Reference Commands

```bash
# Build and push manually
docker build -t ghcr.io/USER/roadmap_tool:latest .
docker push ghcr.io/USER/roadmap_tool:latest

# Create k3d cluster
k3d cluster create roadmap-cluster

# Create image pull secret
kubectl create secret docker-registry ghcr-secret \
  --docker-server=ghcr.io \
  --docker-username=USER \
  --docker-password=PAT

# Deploy
kubectl apply -f deployments/kubernetes/

# Port forward
kubectl port-forward svc/roadmap-visualizer 8080:8080

# Check logs
kubectl logs -l app=roadmap-visualizer -f

# Delete cluster
k3d cluster delete roadmap-cluster
```
