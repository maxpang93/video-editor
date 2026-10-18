# video-editor

## Cluster Setup
Generate K3S_TOKEN and put in .env
```bash
# Option 1: 32-byte hex string (recommended)
openssl rand -hex 32

# Option 2: UUID
uuidgen
```

Start the cluster:
```bash
docker compose up -d
```

Verify joined nodes inside K3s:
```bash
docker exec -it k3s-master kubectl get nodes
```

## Local Image Registry Setup
Check local image registry reachable inside k3s nodes:
```bash
docker compose exec k3s-worker-lite curl -i http://registry:5000/v2/
```

Build and push ffmpeg worker image to image registry
```bash
docker build -t localhost:5000/ffmpeg-worker:latest backend/ -f backend/Dockerfile.worker
docker push localhost:5000/ffmpeg-worker:latest
```

Verify can pull image
```bash
docker compose exec k3s-worker-lite k3s crictl pull registry:5000/ffmpeg-worker:latest
```

## Argo Workflow
Setup
```bash
export KUBECONFIG=.k3s-kubeconfig/k3s.yaml && \
kubectl create namespace argo && \
kubectl apply --server-side -n argo \
  -f https://github.com/argoproj/argo-workflows/releases/download/v4.1.5/install.yaml && \
kubectl -n argo apply -f deploy/argo/trim-merge.yaml
```

Check workflow status
```bash
kubectl -n argo get workflow
```

Observe workflow execution's pod lifecycle
```bash
kubectl -n argo get pods -w
```
