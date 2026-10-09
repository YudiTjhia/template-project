# Encryption Package

Requests use a hybrid envelope: AES-256-GCM encrypts the JSON body and RSA-OAEP-SHA256 wraps the AES key. Generate independent keys for customer and cashier services.

```sh
cd backend
umask 077
go run ./cmd/encryption-keygen -name customer -out-dir /secure/pos-keys
go run ./cmd/encryption-keygen -name cashier -out-dir /secure/pos-keys
base64 -w0 /secure/pos-keys/customer_encryption_private.pem
base64 -w0 /secure/pos-keys/cashier_encryption_private.pem
```

Store each base64 value in Secret Manager or a local ignored `.env` file as `CUSTOMER_ENCRYPTION_PRIVATE_KEY_B64` and `CASHIER_ENCRYPTION_PRIVATE_KEY_B64`. The backend only reads the role-specific private key. Never put private key files or their values in source control, Flutter assets, or browser environment variables.

The cashier deployment script forwards `CASHIER_ENCRYPTION_PRIVATE_KEY_B64` as a runtime secret; the customer deployment script does the same for `CUSTOMER_ENCRYPTION_PRIVATE_KEY_B64`. Local Docker Compose reads the matching value from its `.env` file.

All replicas behind one customer or cashier service must use the same role-specific private key so a request encrypted for one replica can be decrypted by another.

When configured, each backend exposes `GET /.well-known/request-encryption-key`. The customer app encrypts JSON request bodies directly. The cashier Next.js API proxy encrypts JSON bodies before forwarding them to the Go backend. Requests without a JSON body, including GET and file upload requests, are unchanged. If the private key is not configured, the key endpoint is unavailable and encrypted requests cannot proceed.

Use HTTPS in production. Encryption does not replace TLS, session authentication, or authorization; TLS authenticates the key endpoint and protects response bodies and metadata.
