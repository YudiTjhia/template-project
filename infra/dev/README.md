# Development services

Gunakan `build-services.sh` untuk mengelola frontend dan backend sekaligus. Script menjalankan Compose frontend dan backend secara berurutan; masing-masing memakai file `.env` di direktorinya sendiri.

```bash
# Build image frontend dan backend.
infra/dev/build-services.sh build

# Build lalu jalankan keduanya.
infra/dev/build-services up
```

Untuk menghentikan keduanya, jalankan `infra/dev/build-services.sh down`. Masing-masing `docker-build.sh` menerima perintah `down`, `build`, atau `up`; `up` menjalankan container dari image yang sudah dibuild. Database tetap dikelola terpisah melalui `infra/dev/db/docker-build.sh`.
