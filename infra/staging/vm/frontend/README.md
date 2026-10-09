# Frontend staging VM

Frontend Compose ini terpisah dari backend Compose di `infra/staging/vm/backend`. Isi `SUB_APP_TYPE` dan URL API di `.env`, lalu jalankan frontend:

```bash
infra/staging/vm/frontend/docker-build.sh up
```

Frontend tersedia pada port `32006` secara default. URL API harus dapat dijangkau browser staging.
