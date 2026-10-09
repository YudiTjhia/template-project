import 'dart:convert';

import 'package:mobile_app/config/app_config.dart';
import 'package:mobile_app/models/account_profile_model.dart';
import 'package:mobile_app/shared/helpers/session_helper.dart';
import 'package:mobile_app/shared/services/base_service.dart';

class AccountProfileService extends BaseService {
  AccountProfileService({super.client});

  Future<AccountProfileModel> getCurrentAccount() async {
    final accountId = await SessionHelper.getAccountId();
    if (accountId == null) {
      throw const AccountProfileServiceException(
        'Sesi account tidak ditemukan',
      );
    }

    try {
      final response = await client.get(
        Uri.parse('${AppConfig.baseUrl}/api/accounts/$accountId'),
        headers: await sessionHeaders,
      );
      final decoded = jsonDecode(response.body);
      if (decoded is! Map) {
        throw const AccountProfileServiceException(
          'Format respons account tidak valid',
        );
      }
      final payload = Map<String, dynamic>.from(decoded);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw AccountProfileServiceException(
          payload['message']?.toString() ??
              payload['error']?.toString() ??
              'Gagal memuat account (${response.statusCode})',
        );
      }
      final data = payload['data'];
      if (data is! Map) {
        throw const AccountProfileServiceException('Data account tidak valid');
      }
      return AccountProfileModel.fromJson(Map<String, dynamic>.from(data));
    } on AccountProfileServiceException {
      rethrow;
    } on Exception catch (error) {
      throw AccountProfileServiceException(
        'Tidak dapat memuat account: $error',
      );
    }
  }
}

class AccountProfileServiceException implements Exception {
  const AccountProfileServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}
