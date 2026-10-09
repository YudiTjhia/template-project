import 'dart:convert';

import 'package:mobile_app/config/app_config.dart';
import 'package:http/http.dart' as http;

import '../models/login_model.dart';

import 'package:mobile_app/shared/services/base_service.dart';
import 'package:mobile_app/shared/helpers/session_helper.dart';

class LoginServiceException implements Exception {
  const LoginServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}

class LoginService extends BaseService {
  LoginService({super.client});

  Future<void> login(LoginModel model) async {
    final uri = Uri.parse('${AppConfig.baseUrl}/api/auth/login');
    final previousToken = await SessionHelper.getToken();
    late final http.Response response;

    try {
      response = await client.post(
        uri,
        headers: {
          ...await jsonHeaders,
          if (previousToken != null && previousToken.isNotEmpty)
            'X-Session-Token': previousToken,
        },
        body: await encryptJsonBody({...model.toJson()}),
      );
    } on Exception catch (error) {
      throw LoginServiceException('Tidak dapat terhubung ke server: $error');
    }

    late final Map<String, dynamic> payload;
    try {
      final decoded = jsonDecode(response.body);
      if (decoded is! Map) {
        throw const FormatException('Invalid response object');
      }
      payload = Map<String, dynamic>.from(decoded);
    } on FormatException {
      throw const LoginServiceException('Format respons server tidak valid');
    }

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw LoginServiceException(
        payload['message']?.toString() ??
            payload['error']?.toString() ??
            'Login gagal (${response.statusCode})',
      );
    }

    final token = payload['token']?.toString().trim();
    if (token == null || token.isEmpty) {
      throw const LoginServiceException(
        'Server tidak mengembalikan token autentikasi',
      );
    }
    final accountId = _accountIdFrom(payload);
    final userId = _userIdFrom(payload);
    if (accountId == null || accountId <= 0 || userId == null || userId <= 0) {
      throw const LoginServiceException(
        'Respons login tidak memiliki account_id dan user_id yang valid',
      );
    }

    try {
      await SessionHelper.saveLoginSession(
        token: token,
        accountId: accountId,
        userId: userId,
      );
    } on StateError catch (error) {
      throw LoginServiceException(error.message);
    }
  }

  Future<void> loginWithGoogleToken(String idToken) async {
    final previousToken = await SessionHelper.getToken();
    late final http.Response response;
    try {
      response = await client.post(
        Uri.parse('${AppConfig.baseUrl}/api/auth/google'),
        headers: {
          ...await jsonHeaders,
          if (previousToken != null && previousToken.isNotEmpty)
            'X-Session-Token': previousToken,
        },
        body: await encryptJsonBody({'idToken': idToken}),
      );
    } on Exception catch (error) {
      throw LoginServiceException('Tidak dapat terhubung ke server: $error');
    }

    final payload = _decodeObject(response.body);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw LoginServiceException(
        payload['message']?.toString() ??
            payload['error']?.toString() ??
            'Login dengan Google gagal (${response.statusCode})',
      );
    }

    final token = payload['token']?.toString().trim();
    final accountId = _accountIdFrom(payload);
    final userId = _userIdFrom(payload);
    if (token == null ||
        token.isEmpty ||
        accountId == null ||
        accountId <= 0 ||
        userId == null ||
        userId <= 0) {
      throw const LoginServiceException('Respons login Google tidak valid');
    }
    try {
      await SessionHelper.saveLoginSession(
        token: token,
        accountId: accountId,
        userId: userId,
      );
    } on StateError catch (error) {
      throw LoginServiceException(error.message);
    }
  }

  Future<bool> hasSession() async {
    return SessionHelper.hasLoginSession();
  }

  Future<bool> validateSession() async {
    final token = await SessionHelper.getToken();
    final accountId = await SessionHelper.getAccountId();
    final userId = await SessionHelper.getUserId();
    if (token == null ||
        token.isEmpty ||
        accountId == null ||
        accountId <= 0 ||
        userId == null ||
        userId <= 0) {
      await clearSession();
      return false;
    }

    late final http.Response response;
    try {
      response = await client.post(
        Uri.parse('${AppConfig.baseUrl}/api/auth/refresh'),
        headers: {
          ...await headers,
          'X-Session-Token': token,
          'Accept': 'application/json',
          'Authorization': 'Bearer $token',
        },
      );
    } on Exception catch (error) {
      throw LoginServiceException('Tidak dapat memvalidasi sesi: $error');
    }

    if (response.statusCode == 401) {
      await clearSession();
      return false;
    }
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw LoginServiceException(
        _decodeMessage(response.body) ??
            'Gagal memvalidasi sesi (${response.statusCode})',
      );
    }

    final payload = _decodeObject(response.body);
    if (payload['ok'] != true) {
      await clearSession();
      return false;
    }

    final validatedAccountId = _accountIdFrom(payload) ?? accountId;
    final validatedUserId = _userIdFrom(payload) ?? userId;
    if (validatedAccountId <= 0 || validatedUserId <= 0) {
      await clearSession();
      return false;
    }

    try {
      await SessionHelper.saveLoginSession(
        token: payload['token']?.toString().trim().isNotEmpty == true
            ? payload['token'].toString().trim()
            : token,
        accountId: validatedAccountId,
        userId: validatedUserId,
      );
    } on StateError catch (error) {
      throw LoginServiceException(error.message);
    }
    return true;
  }

  Future<void> clearSession() async {
    await SessionHelper.clearLoginSession();
  }

  Future<void> logout() async {
    final token = await SessionHelper.getToken();
    try {
      if (token != null) {
        await client.post(
          Uri.parse('${AppConfig.baseUrl}/api/auth/logout'),
          headers: {
            ...await headers,
            'X-Session-Token': token,
            'Authorization': 'Bearer $token',
          },
        );
      }
    } finally {
      await clearSession();
    }
  }

  Future<void> changePassword({
    required String currentPassword,
    required String newPassword,
    required String confirmPassword,
  }) async {
    late final http.Response response;
    try {
      response = await client.post(
        Uri.parse('${AppConfig.baseUrl}/api/auth/change-password-pin'),
        headers: await sessionJsonHeaders,
        body: await encryptJsonBody({
          'currentPassword': currentPassword,
          'newPassword': newPassword,
          'confirmPassword': confirmPassword,
          'pin': '',
          'confirmPin': '',
        }),
      );
    } on Exception catch (error) {
      throw LoginServiceException('Tidak dapat mengubah password: $error');
    }

    final payload = _decodeObject(response.body);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw LoginServiceException(
        payload['message']?.toString() ??
            payload['error']?.toString() ??
            'Gagal mengubah password (${response.statusCode})',
      );
    }
  }

  Future<int?> getAccountId() => SessionHelper.getAccountId();

  Future<int?> getUserId() => SessionHelper.getUserId();

  Map<String, dynamic> _decodeObject(String body) {
    try {
      final decoded = jsonDecode(body);
      if (decoded is Map) return Map<String, dynamic>.from(decoded);
    } on FormatException {
      throw const LoginServiceException('Format respons server tidak valid');
    }
    throw const LoginServiceException('Format respons server tidak valid');
  }

  String? _decodeMessage(String body) {
    if (body.isEmpty) return null;
    final payload = _decodeObject(body);
    return payload['message']?.toString() ?? payload['error']?.toString();
  }

  int? _accountIdFrom(Map<String, dynamic> payload) {
    final user = payload['user'];
    final value =
        payload['account_id'] ?? (user is Map ? user['account_id'] : null);
    return _parseId(value);
  }

  int? _userIdFrom(Map<String, dynamic> payload) {
    final user = payload['user'];
    final value =
        payload['user_id'] ??
        (user is Map ? user['user_id'] ?? user['id'] : null);
    return _parseId(value);
  }

  int? _parseId(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    return int.tryParse(value?.toString() ?? '');
  }
}
