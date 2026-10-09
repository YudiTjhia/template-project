# Backend dev

Backend, frontend, dan database memiliki Compose terpisah. Nyalakan database terlebih dahulu, lalu backend:

```bash
# Pastikan kedua file .env sudah diisi.
infra/dev/db/docker-build.sh up
infra/dev/backend/docker-build.sh up
```

Backend tersedia pada port `31007` secara default dan health endpoint di `/healthz`. Compose backend dan database berbagi network `template_project_dev`; kedua DSN diarahkan ke host `db`.
