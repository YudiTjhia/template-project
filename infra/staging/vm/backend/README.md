# Backend staging VM

Compose backend ini berdiri sendiri dari frontend. Pastikan `.env` telah diisi, lalu jalankan:

```bash
infra/staging/vm/backend/docker-build.sh up
```

Backend tersedia pada port `32007` secara default dan health endpoint di `/healthz`.
