# Frontend Web

Aplikasi web Next.js 15 dengan React 19 dan TypeScript. Source halaman, API routes, komponen, serta service berada di `app/`.

## Menjalankan lokal

Dari direktori `frontend/web/`:

```bash
npm ci
npm run dev
```

Buka `http://localhost:3000`. Buat file `.env` lokal (diabaikan Git) dan isi setidaknya URL backend serta jenis sub-app:

```dotenv
SUB_APP_TYPE=admin
ADMIN_API_BASE_URL=http://localhost:8080
CUSTOMER_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

`SUB_APP_TYPE` memilih `admin` atau `customer` dan menentukan prefix route yang terlihat (`/admin/...` atau `/customer/...`). Variabel `NEXT_PUBLIC_*` tertanam di build browser; jangan taruh credential privat di dalamnya. Untuk menjalankan web, lihat juga [panduan infra development](../../infra/dev/frontend/README.md).

## Build

```bash
npm run build
npm run start
```

Container production dibangun dengan `Dockerfile`. Konfigurasi environment untuk Compose development, VM staging, dan Cloud Run dijelaskan di [README infra](../../infra/README.md) serta README masing-masing deployment.

## Firebase Web Messaging

Set these values in `frontend/web/.env` for local Next.js development. For Docker and Cloud Run builds, set the same values in the matching `infra/**/frontend/.env`; the build configuration passes them into Next.js.

```dotenv
NEXT_PUBLIC_FIREBASE_API_KEY=
NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN=
NEXT_PUBLIC_FIREBASE_PROJECT_ID=
NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET=
NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID=
NEXT_PUBLIC_FIREBASE_APP_ID=
NEXT_PUBLIC_FIREBASE_VAPID_KEY=
```

Get the web app values from Firebase Console → Project settings → General → Your apps. Generate a Web Push certificate in Project settings → Cloud Messaging and use its public key for `NEXT_PUBLIC_FIREBASE_VAPID_KEY`. The app serves its messaging worker at `/firebase-messaging-sw.js`.

Firebase web configuration and the VAPID public key are client-side values. Keep Firebase Admin SDK credentials and service-account private keys on the backend only. The frontend can obtain an FCM registration token, but backend token registration is not implemented yet.

The signed-in notification page is available at `/notifications`. It loads the user's notification list and unread count through the same-origin `/api/notifications` routes.
# Sub-app routing

The active `SUB_APP_TYPE` adds a visible route prefix: `/admin/...` for the admin build and `/customer/...` for the customer build. Requests without a prefix are redirected to the active sub-app. The middleware rewrites the prefixed URL internally to the existing route group, so route files remain shared.

The `admin` and `customer` builds are still configured independently with `SUB_APP_TYPE` and their corresponding API URL.
