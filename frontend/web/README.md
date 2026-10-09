# Frontend Web

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
