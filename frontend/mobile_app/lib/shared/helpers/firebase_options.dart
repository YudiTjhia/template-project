import 'package:firebase_core/firebase_core.dart' show FirebaseOptions;
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, kIsWeb, TargetPlatform;

class DefaultFirebaseOptions {
  static FirebaseOptions get currentPlatform {
    if (kIsWeb) {
      throw UnsupportedError('Firebase options are not configured for web.');
    }
    if (defaultTargetPlatform == TargetPlatform.android) return android;
    throw UnsupportedError('Firebase options are only configured for Android.');
  }

  static const FirebaseOptions android = FirebaseOptions(
    apiKey: 'AIzaSyCjca6vAE8DiR4RoLuvt5TnukAnS8UMth0',
    appId: '1:406119505528:android:365c84e25180e4fbd4178d',
    messagingSenderId: '406119505528',
    projectId: 'market-pos-93cc1',
    storageBucket: 'market-pos-93cc1.firebasestorage.app',
  );
}
