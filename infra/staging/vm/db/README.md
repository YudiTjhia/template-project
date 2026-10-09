# PostgreSQL staging VM database

Menjalankan PostgreSQL 16 dengan pgvector, database master dan transaction, healthcheck, serta persistent volume. Port database hanya bind ke `127.0.0.1:32008`; service backend di network `template_web_staging` dapat terhubung melalui host `db`.

```bash
# Ganti POSTGRES_PASSWORD sebelum menjalankan staging.
infra/staging/vm/db/docker-build.sh up
```

File SQL init hanya berjalan saat volume baru dibuat. Simpan backup database terpisah sebelum menghapus volume.
