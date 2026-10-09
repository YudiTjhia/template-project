# Frontend dev

Frontend berjalan terpisah dari backend. Atur `SUB_APP_TYPE` dan API origin di `.env`, lalu jalankan Compose frontend. Backend memiliki Compose dan lifecycle sendiri di `infra/dev/backend`.

```bash
# Pastikan infra/dev/frontend/.env sudah diisi.
infra/dev/frontend/docker-build.sh up
```

Frontend tersedia di `http://localhost:31006` secara default. `.env` berisi konfigurasi lokal dan diabaikan Git; buat atau isi file itu sebelum menjalankan Compose.
