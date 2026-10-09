# Frontend Mobile App

Aplikasi mobile Flutter dengan halaman autentikasi, home, profil, pengguna, role, dan notifikasi. Source aplikasi berada di `lib/`; service API ada di `lib/services/` dan konfigurasi runtime dasar ada di `lib/shared/config/app_config.dart`.

## Persyaratan

- Flutter SDK dan Dart SDK yang memenuhi constraint di `pubspec.yaml`.
- Android Studio/Xcode atau toolchain platform tujuan untuk menjalankan aplikasi di emulator/perangkat.

## Persiapan dan menjalankan

Dari direktori `frontend/mobile_app/`:

```bash
flutter pub get
```

Salin `.env.example` menjadi `.env` lokal (file `.env` tidak disimpan di Git):

```bash
cp .env.example .env
```

Atur konfigurasi di `.env`:

```dotenv
BASE_URL=http://10.0.2.2:8080
APP_SOURCE=customer_app
APP_TITLE=Open POS
USE_DUMMY_USERS=true
USE_DUMMY_ROLES=true
```

`BASE_URL` adalah alamat backend yang dapat dijangkau perangkat. `10.0.2.2` mengarah ke host developer dari Android Emulator; gunakan alamat LAN host untuk perangkat fisik, dan URL yang sesuai untuk iOS Simulator atau target lain. Nilai default di aplikasi adalah `http://127.0.0.1:8080`. `APP_SOURCE` default-nya `customer_app`. `APP_TITLE` digunakan di title aplikasi, splash, login, register, home, dan drawer; nilai default-nya `Open POS`. `USE_DUMMY_USERS=true` menampilkan data contoh dari `DummyUserService`, sedangkan `USE_DUMMY_ROLES=true` memakai `DummyRoleService` untuk daftar, detail, action lookup, dan CRUD role secara in-memory. Ubah masing-masing ke `false` untuk memakai service backend tanpa mengubah halaman.

Jalankan aplikasi:

```bash
flutter run
```

Untuk memilih perangkat, lihat `flutter devices`, lalu jalankan `flutter run -d <device-id>`.

## Build

```bash
flutter build apk
flutter build appbundle
flutter build ios
```

Build iOS memerlukan macOS dan Xcode. Konfigurasi Firebase platform ada di `lib/shared/helpers/firebase_options.dart`; Firebase Messaging dan notifikasi lokal saat ini diinisialisasi pada Android.

## Catatan integrasi

Backend yang berjalan saat ini hanya menyediakan health check dan dua route stub (`POST /account/register`, `POST /auth/login` yang membalas 501). Lihat [README backend](../../backend/README.md) untuk status API dan [README infra](../../infra/README.md) untuk menjalankan layanan lokal.
