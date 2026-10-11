# video-editor

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