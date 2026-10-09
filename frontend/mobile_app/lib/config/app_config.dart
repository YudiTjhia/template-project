import 'dart:math';

import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:shared_preferences/shared_preferences.dart';

class AppConfig {
  static const _clientIdKey = 'clientId';
  static Future<String>? _clientIdFuture;

  static String get baseUrl {
    return dotenv.env['BASE_URL'] ?? 'http://127.0.0.1:8080';
  }

  static String get appTitle {
    final title = dotenv.env['APP_TITLE']?.trim();
    return title == null || title.isEmpty ? 'Open POS' : title;
  }

  static bool get useDummyUsers =>
      (dotenv.env['USE_DUMMY_USERS'] ?? 'false').trim().toLowerCase() == 'true';

  static bool get useDummyRoles =>
      (dotenv.env['USE_DUMMY_ROLES'] ?? 'false').trim().toLowerCase() == 'true';

  static Future<String> get clientId =>
      _clientIdFuture ??= _loadOrCreateClientId();

  static Future<String> _loadOrCreateClientId() async {
    final preferences = await SharedPreferences.getInstance();
    final existing = preferences.getString(_clientIdKey);
    if (existing != null && existing.isNotEmpty) return existing;

    final generated = _generateClientId();
    final saved = await preferences.setString(_clientIdKey, generated);
    if (!saved) throw StateError('Gagal menyimpan client ID instalasi');
    return generated;
  }

  static String get appSource => dotenv.env['APP_SOURCE'] ?? 'customer_app';

  static String _generateClientId() {
    final random = Random.secure();
    final bytes = List<int>.generate(16, (_) => random.nextInt(256));
    bytes[6] = (bytes[6] & 0x0f) | 0x40;
    bytes[8] = (bytes[8] & 0x3f) | 0x80;
    final hex = bytes.map((byte) => byte.toRadixString(16).padLeft(2, '0'));
    final value = hex.join();
    return '${value.substring(0, 8)}-${value.substring(8, 12)}-'
        '${value.substring(12, 16)}-${value.substring(16, 20)}-'
        '${value.substring(20)}';
  }
}
