import 'package:shared_preferences/shared_preferences.dart';

class SessionHelper {
  SessionHelper._();

  static const _tokenKey = 'authToken';
  static const _accountIdKey = 'account_id';
  static const _userIdKey = 'user_id';

  static Future<bool> hasLoginSession() async {
    final preferences = await SharedPreferences.getInstance();
    final token = preferences.getString(_tokenKey);
    final accountId = preferences.getInt(_accountIdKey);
    final userId = preferences.getInt(_userIdKey);
    return token != null &&
        token.isNotEmpty &&
        accountId != null &&
        accountId > 0 &&
        userId != null &&
        userId > 0;
  }

  static Future<void> saveLoginSession({
    required String token,
    required int accountId,
    required int userId,
  }) async {
    if (token.trim().isEmpty || accountId <= 0 || userId <= 0) {
      throw ArgumentError('Token, account ID, dan user ID harus valid');
    }

    final preferences = await SharedPreferences.getInstance();
    final tokenSaved = await preferences.setString(_tokenKey, token.trim());
    if (!tokenSaved) throw StateError('Gagal menyimpan token sesi');

    final accountIdSaved = await preferences.setInt(_accountIdKey, accountId);
    final userIdSaved = await preferences.setInt(_userIdKey, userId);
    if (!accountIdSaved || !userIdSaved) {
      await clearLoginSession();
      throw StateError('Gagal menyimpan account ID dan user ID sesi');
    }
  }

  static Future<String?> getToken() async {
    final preferences = await SharedPreferences.getInstance();
    final token = preferences.getString(_tokenKey);
    return token == null || token.isEmpty ? null : token;
  }

  static Future<int?> getAccountId() async {
    final preferences = await SharedPreferences.getInstance();
    final accountId = preferences.getInt(_accountIdKey);
    return accountId != null && accountId > 0 ? accountId : null;
  }

  static Future<int?> getUserId() async {
    final preferences = await SharedPreferences.getInstance();
    final userId = preferences.getInt(_userIdKey);
    return userId != null && userId > 0 ? userId : null;
  }

  static Future<void> clearLoginSession() async {
    final preferences = await SharedPreferences.getInstance();
    await preferences.remove(_tokenKey);
    await preferences.remove(_accountIdKey);
    await preferences.remove(_userIdKey);
  }
}
