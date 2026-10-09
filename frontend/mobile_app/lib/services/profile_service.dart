import 'dart:convert';

import 'package:mobile_app/config/app_config.dart';
import 'package:mobile_app/shared/helpers/session_helper.dart';
import 'package:mobile_app/shared/services/base_service.dart';

import 'package:mobile_app/models/profile_model.dart';

class ProfileService extends BaseService {
  ProfileService({super.client});

  Future<ProfileModel> getCurrentUser() async {
    final userId = await SessionHelper.getUserId();
    if (userId == null) {
      throw const ProfileServiceException('Sesi user tidak ditemukan');
    }

    try {
      final response = await client.get(
        Uri.parse('${AppConfig.baseUrl}/api/users/$userId'),
        headers: await sessionHeaders,
      );

      Map<String, dynamic> payload;
      try {
        final decoded = jsonDecode(response.body);
        if (decoded is! Map) throw const FormatException();
        payload = Map<String, dynamic>.from(decoded);
      } on FormatException {
        throw const ProfileServiceException(
          'Format respons profil tidak valid',
        );
      }

      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw ProfileServiceException(
          payload['message']?.toString() ??
              payload['error']?.toString() ??
              'Gagal mengambil profil (${response.statusCode})',
        );
      }

      final data = payload['data'];
      if (data is! Map) {
        throw const ProfileServiceException('Data profil tidak valid');
      }
      return ProfileModel.fromJson(Map<String, dynamic>.from(data));
    } on ProfileServiceException {
      rethrow;
    } on Exception catch (error) {
      throw ProfileServiceException('Tidak dapat memuat profil: $error');
    }
  }

  Future<void> updateCurrentUser({
    required String username,
    required String email,
  }) async {
    final userId = await SessionHelper.getUserId();
    if (userId == null) {
      throw const ProfileServiceException('Sesi user tidak ditemukan');
    }

    try {
      final response = await client.patch(
        Uri.parse('${AppConfig.baseUrl}/api/users/$userId/profile'),
        headers: await sessionJsonHeaders,
        body: await encryptJsonBody({
          'username': username.trim(),
          'email': email.trim(),
        }),
      );

      if (response.statusCode < 200 || response.statusCode >= 300) {
        final payload = _decodeObject(response.body);
        throw ProfileServiceException(
          payload['message']?.toString() ??
              payload['error']?.toString() ??
              'Gagal memperbarui profil (${response.statusCode})',
        );
      }
    } on ProfileServiceException {
      rethrow;
    } on Exception catch (error) {
      throw ProfileServiceException('Tidak dapat memperbarui profil: $error');
    }
  }

  Map<String, dynamic> _decodeObject(String body) {
    try {
      final decoded = jsonDecode(body);
      if (decoded is Map) return Map<String, dynamic>.from(decoded);
    } on FormatException {
      throw const ProfileServiceException('Format respons profil tidak valid');
    }
    throw const ProfileServiceException('Format respons profil tidak valid');
  }
}

class ProfileServiceException implements Exception {
  const ProfileServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}
