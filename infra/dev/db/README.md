# PostgreSQL dev database

Menjalankan PostgreSQL 16 dengan pgvector, dua database (`template_web_master` dan `template_web_tr`), serta persistent volume. Backend dan database berbagi Docker network `template_web_dev`; DSN backend memakai host `db`.

```bash
# Pastikan infra/dev/db/.env sudah diisi.
infra/dev/db/docker-build.sh up
```

Port host default `31008` hanya bind ke localhost. File SQL init berjalan hanya saat volume pertama kali dibuat.
