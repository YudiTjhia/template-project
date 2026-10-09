import 'dart:convert';

import 'package:http/http.dart' as http;

class ApiException implements Exception {
  const ApiException(this.message);

  final String message;

  @override
  String toString() => message;
}

class ApiHelper {
  static Map<String, dynamic> decodeApiObject(String body) {
    if (body.isEmpty) return {};

    final decoded = jsonDecode(body);
    if (decoded is Map) {
      return Map<String, dynamic>.from(decoded);
    }

    throw const ApiException('Format respons server tidak valid');
  }

  static void ensureApiSuccess(
    http.Response response,
    Map<String, dynamic> payload,
    String fallbackMessage,
  ) {
    if (response.statusCode >= 200 && response.statusCode < 300) return;

    throw ApiException(
      payload['message']?.toString() ??
          payload['error']?.toString() ??
          '$fallbackMessage (${response.statusCode})',
    );
  }

  static List<Map<String, dynamic>> requireApiList(
    Map<String, dynamic> payload,
  ) {
    final data = payload['data'];
    if (data is! List) {
      throw const ApiException('Format daftar data dari server tidak valid');
    }
    return data
        .map((item) {
          if (item is Map) return Map<String, dynamic>.from(item);
          throw const ApiException('Format item data dari server tidak valid');
        })
        .toList(growable: false);
  }

  static Map<String, dynamic> requireApiData(Map<String, dynamic> payload) {
    final data = payload['data'];
    if (data is Map) return Map<String, dynamic>.from(data);
    throw const ApiException('Format detail data dari server tidak valid');
  }

  static int requireApiId(Map<String, dynamic> payload) {
    final value = payload['id'];
    if (value is int) return value;
    if (value is num) return value.toInt();
    final parsed = int.tryParse(value?.toString() ?? '');
    if (parsed != null) return parsed;
    throw const ApiException('Server tidak mengembalikan ID yang valid');
  }
}
