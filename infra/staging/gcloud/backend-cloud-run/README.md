# Backend staging Cloud Run

Deploy backend secara mandiri. Image dibangun dari `backend/`; endpoint health tersedia di `/healthz`.

```bash
# Isi file .env dengan project, service, dan runtime service account.
infra/staging/gcloud/deploy-backend-cloud-run.sh --check
infra/staging/gcloud/deploy-backend-cloud-run.sh
```

API `/account/register` dan `/auth/login` masih stub `501 Not Implemented`.
