import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/config/app_config.dart';
import 'package:mobile_app/shared/services/base_service.dart';

class UserService extends BaseService {
  UserService({super.client});

  Future<List<UserModel>> getUsers({String search = ''}) async {
    final uri = Uri.parse('${AppConfig.baseUrl}/api/users').replace(
      queryParameters: {if (search.trim().isNotEmpty) 'search': search.trim()},
    );
    final payload = await _requestObject(
      () async => client.get(uri, headers: await sessionHeaders),
      'Gagal memuat daftar user',
    );
    final data = payload['data'];
    if (data is! List) {
      throw const UserServiceException('Data user tidak valid');
    }
    return data
        .whereType<Map>()
        .map((item) => UserModel.fromJson(Map<String, dynamic>.from(item)))
        .toList(growable: false);
  }

  Future<UserModel> getUser(int id) async {
    final payload = await _requestObject(
      () async => client.get(
        Uri.parse('${AppConfig.baseUrl}/api/users/$id'),
        headers: await sessionHeaders,
      ),
      'Gagal memuat user',
    );
    final data = payload['data'];
    if (data is! Map) {
      throw const UserServiceException('Data user tidak valid');
    }
    return UserModel.fromJson(Map<String, dynamic>.from(data));
  }

  Future<void> createUser(
    UserModel user, {
    required String password,
    String pin = '',
  }) async {
    await _requestObject(
      () async => client.post(
        Uri.parse('${AppConfig.baseUrl}/api/users'),
        headers: await sessionJsonHeaders,
        body: await encryptJsonBody(user.toJson(password: password, pin: pin)),
      ),
      'Gagal menyimpan user',
    );
  }

  Future<void> updateUser(
    UserModel user, {
    required String password,
    String pin = '',
  }) async {
    await _requestObject(
      () async => client.put(
        Uri.parse('${AppConfig.baseUrl}/api/users/${user.id}'),
        headers: await sessionJsonHeaders,
        body: await encryptJsonBody(user.toJson(password: password, pin: pin)),
      ),
      'Gagal memperbarui user',
    );
  }

  Future<void> deleteUser(int id) async {
    await _requestObject(
      () async => client.delete(
        Uri.parse('${AppConfig.baseUrl}/api/users/$id'),
        headers: await sessionHeaders,
      ),
      'Gagal menghapus user',
    );
  }

  Future<Map<String, dynamic>> _requestObject(
    Future<http.Response> Function() request,
    String fallback,
  ) async {
    try {
      final response = await request();
      final decoded = jsonDecode(response.body);
      if (decoded is! Map) {
        throw const UserServiceException('Format respons user tidak valid');
      }
      final payload = Map<String, dynamic>.from(decoded);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw UserServiceException(
          payload['message']?.toString() ??
              payload['error']?.toString() ??
              '$fallback (${response.statusCode})',
        );
      }
      return payload;
    } on UserServiceException {
      rethrow;
    } on Exception catch (error) {
      throw UserServiceException('$fallback: $error');
    }
  }
}

class UserServiceException implements Exception {
  const UserServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}
