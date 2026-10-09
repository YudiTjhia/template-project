import 'dart:convert';

import 'package:mobile_app/shared/config/app_config.dart';
import 'package:mobile_app/shared/helpers/session_helper.dart';
import 'package:mobile_app/shared/services/base_service.dart';

import '../models/profile_model.dart';

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
}

class ProfileServiceException implements Exception {
  const ProfileServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}
