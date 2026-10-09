# Backend

Backend HTTP berbasis Go. Entry point berada di `main.go`; kode bersama untuk model, service, middleware, dan route tersedia di `shared/` sebagai modul lokal `backend/shared`.

## Persyaratan

- Go sesuai versi pada `go.mod` (Go 1.26.3).
- Docker, jika ingin menjalankan backend dalam container.

## Menjalankan lokal

Dari direktori `backend/`:

```bash
go run .
```

Server mendengarkan pada `http://localhost:8080` secara default. Port dapat diubah melalui `PORT`; `CORS_ALLOW_ORIGIN` mengatur satu origin yang diizinkan dan default-nya `*`.

```bash
PORT=8081 CORS_ALLOW_ORIGIN=http://localhost:3000 go run .
```

Endpoint yang tersedia pada entry point saat ini:

| Method | Path | Keterangan |
| --- | --- | --- |
| `GET` | `/healthz` | Health check, merespons JSON `{"status":"ok"}`. |
| `POST` | `/account/register` | Belum diimplementasikan; merespons HTTP 501. |
| `POST` | `/auth/login` | Belum diimplementasikan; merespons HTTP 501. |

Path lain merespons 404. Route dan service tambahan di `shared/` belum otomatis didaftarkan oleh entry point ini.

## Docker

Dari root repository:

```bash
docker build -t template-project-backend ./backend
docker run --rm -p 8080:8080 template-project-backend
```

Konfigurasi Docker Compose development dan staging ada di [infra](../infra/README.md). Untuk konfigurasi database, lihat README pada direktori environment tersebut.

## Struktur

- `main.go` — HTTP server, CORS, health check, dan route entry point.
- `shared/` — modul bersama yang diimpor melalui `replace` lokal di `go.mod`.
- `Dockerfile` — build multi-stage untuk image backend.
