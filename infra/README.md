# Infrastructure

Direktori ini berisi konfigurasi Docker Compose dan script untuk menjalankan layanan development serta menyiapkan deployment staging. Backend, frontend web, dan database dikelola terpisah agar dapat dinyalakan atau dibangun sesuai kebutuhan.

## Pilihan environment

| Environment | Backend | Frontend web | Database |
| --- | --- | --- | --- |
| Development lokal | [`dev/backend`](dev/backend/README.md) | [`dev/frontend`](dev/frontend/README.md) | [`dev/db`](dev/db/README.md) |
| Staging VM | [`staging/vm/backend`](staging/vm/backend/README.md) | [`staging/vm/frontend`](staging/vm/frontend/README.md) | [`staging/vm/db`](staging/vm/db/README.md) |
| Staging Google Cloud Run | [`staging/gcloud/backend-cloud-run`](staging/gcloud/backend-cloud-run/README.md) | [`staging/gcloud/frontend-cloud-run`](staging/gcloud/frontend-cloud-run/README.md) | Dikelola di luar konfigurasi Cloud Run pada repository ini. |

Setiap direktori deployment memiliki README sendiri yang menjelaskan variabel environment, port, dan perintah khusus.

## Development lokal dengan Compose

Isi `.env` pada `infra/dev/db`, `infra/dev/backend`, dan `infra/dev/frontend` terlebih dahulu. File `.env` tidak disimpan di Git; contoh variabel wajib dapat dilihat di file `docker-compose.yml` masing-masing.

Jalankan database, backend, kemudian frontend:

```bash
infra/dev/db/docker-build.sh up
infra/dev/backend/docker-build.sh up
infra/dev/frontend/docker-build.sh up
```

Port default: web `31006`, backend `31007`, database `31008` (database hanya bind ke localhost). DB dan backend berbagi Docker network `template_project_dev`; database diakses dengan hostname `db`. Perintah umum pada tiap wrapper: `up`, `down`, `logs`, `ps`; backend dan frontend juga menerima `build`.

## Staging VM

Isi file `.env` pada `infra/staging/vm/db`, `infra/staging/vm/backend`, dan `infra/staging/vm/frontend`, lalu jalankan Compose berurutan. Petunjuk lengkap ada di [panduan VM](staging/vm/README.md). Port default-nya web `32006`, backend `32007`, dan DB `32008` yang hanya bind ke localhost.

Script `build-db.sh`, `build-backend.sh`, dan `build-frontend.sh` di `staging/vm/` hanya membungkus perintah Compose `up`/`build` untuk layanan terkait; untuk menghentikan, melihat log, atau memeriksa status gunakan `docker-build.sh` di direktori layanan.

## Google Cloud Run

Staging Cloud Run memiliki script deploy di `infra/staging/gcloud/`. Isi file konfigurasi `.env` terkait, periksa konfigurasi dengan opsi `--check`, lalu deploy backend sebelum frontend. Lihat README [backend](staging/gcloud/backend-cloud-run/README.md) dan [frontend](staging/gcloud/frontend-cloud-run/README.md) untuk prasyarat serta langkah detail.

## Konfigurasi dan data

- File `.env`, credential, dan secret tidak boleh dimasukkan ke Git. Gunakan nilai konfigurasi terpisah untuk setiap environment.
- Frontend web membutuhkan URL API yang dapat dijangkau browser. Untuk pengaturan `SUB_APP_TYPE`, Firebase Web Messaging, dan API proxy, lihat [README frontend web](../frontend/web/README.md).
- PostgreSQL Compose menggunakan volume persisten. Script init SQL berjalan hanya saat volume pertama kali dibuat; lihat README DB terkait untuk detail.
- Build mobile app Flutter tidak dikelola oleh Compose infra ini. Petunjuk konfigurasi dan build ada di [README mobile app](../frontend/mobile_app/README.md).
