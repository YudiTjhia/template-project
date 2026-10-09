# Staging VM

Database, backend, dan frontend memiliki Compose serta environment file masing-masing. Jalankan database terlebih dahulu, lalu backend dan frontend:

```bash
# Pastikan .env di db, backend, dan frontend sudah diisi dengan nilai staging.
infra/staging/vm/db/docker-build.sh up
infra/staging/vm/backend/docker-build.sh up
infra/staging/vm/frontend/docker-build.sh up
```

Database tersedia pada port localhost `32008`, backend pada `32007`, dan frontend pada `32006`. Backend dan DB berbagi network `template_project_staging`, dengan host database `db`.
File `.env` berisi konfigurasi sensitif dan diabaikan Git.
