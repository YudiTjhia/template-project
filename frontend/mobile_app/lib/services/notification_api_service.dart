import 'dart:convert';

import 'package:mobile_app/models/notification_model.dart';
import 'package:mobile_app/shared/config/app_config.dart';
import 'package:mobile_app/shared/services/base_service.dart';

class NotificationApiService extends BaseService {
  NotificationApiService({super.client});

  Future<List<NotificationModel>> getNotifications({int limit = 50}) async {
    try {
      final uri = Uri.parse('${AppConfig.baseUrl}/api/notifications')
          .replace(queryParameters: {'limit': '$limit'});
      final response = await client.get(uri, headers: await sessionHeaders);
      final payload = _decodeObject(response.body);
      _ensureSuccess(response.statusCode, payload, 'Gagal memuat notifikasi');
      final data = payload['data'];
      if (data is! List) {
        throw const NotificationApiException('Data notifikasi tidak valid');
      }
      return data
          .map((item) {
            if (item is! Map) {
              throw const NotificationApiException(
                'Format item notifikasi tidak valid',
              );
            }
            return NotificationModel.fromJson(Map<String, dynamic>.from(item));
          })
          .toList(growable: false);
    } on NotificationApiException {
      rethrow;
    } on Exception catch (error) {
      throw NotificationApiException('Tidak dapat memuat notifikasi: $error');
    }
  }

  Future<void> markAsRead(int id) async {
    try {
      final response = await client.post(
        Uri.parse('${AppConfig.baseUrl}/api/notifications/$id/read'),
        headers: await sessionHeaders,
      );
      final payload = _decodeObject(response.body);
      _ensureSuccess(
        response.statusCode,
        payload,
        'Gagal memperbarui notifikasi',
      );
    } on NotificationApiException {
      rethrow;
    } on Exception catch (error) {
      throw NotificationApiException(
        'Tidak dapat memperbarui notifikasi: $error',
      );
    }
  }

  Map<String, dynamic> _decodeObject(String body) {
    try {
      final decoded = jsonDecode(body);
      if (decoded is Map) return Map<String, dynamic>.from(decoded);
    } on FormatException {
      throw const NotificationApiException(
        'Format respons notifikasi tidak valid',
      );
    }
    throw const NotificationApiException(
      'Format respons notifikasi tidak valid',
    );
  }

  void _ensureSuccess(
    int statusCode,
    Map<String, dynamic> payload,
    String fallback,
  ) {
    if (statusCode >= 200 && statusCode < 300) return;
    throw NotificationApiException(
      payload['message']?.toString() ??
          payload['error']?.toString() ??
          '$fallback ($statusCode)',
    );
  }
}

class NotificationApiException implements Exception {
  const NotificationApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
