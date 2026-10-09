# Frontend staging Cloud Run

Deploy frontend secara mandiri setelah backend tersedia. Frontend menggunakan URL dari `BACKEND_SERVICE`, atau API URL override dari file `.env`.

```bash
# Isi file .env dengan project, service, dan runtime service account. Deploy backend terlebih dahulu.
infra/staging/gcloud/deploy-frontend-cloud-run.sh --check
infra/staging/gcloud/deploy-frontend-cloud-run.sh
```

Sub-app dipilih dari `SUB_APP_TYPE` di `frontend/web/.env` (default `admin`), dan bisa dioverride di shell/CI. Pastikan gcloud login dan izin Cloud Build, Artifact Registry, Cloud Run, serta service account tersedia. File `.env` berisi konfigurasi sensitif dan diabaikan Git.
