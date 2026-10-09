import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:mobile_app/models/role_model.dart';
import 'package:mobile_app/shared/config/app_config.dart';
import 'package:mobile_app/shared/services/base_service.dart';

class RoleService extends BaseService {
  RoleService({super.client});

  Future<List<RoleModel>> getRoles() async {
    return _getRoleList(Uri.parse('${AppConfig.baseUrl}/api/roles?all=1'));
  }

  Future<List<RoleActionModel>> getRoleActions() async {
    final payload = await _requestObject(
      () async => client.get(
        Uri.parse('${AppConfig.baseUrl}/api/roles/actions'),
        headers: await sessionHeaders,
      ),
      'Gagal memuat daftar action',
    );
    final data = payload['data'];
    if (data is! List) {
      throw const RoleServiceException('Data action role tidak valid');
    }
    return data
        .whereType<Map>()
        .map(
          (item) => RoleActionModel.fromJson(Map<String, dynamic>.from(item)),
        )
        .where((action) => action.active)
        .toList(growable: false);
  }

  Future<RoleModel> getRole(int id) async {
    final payload = await _requestObject(
      () async => client.get(
        Uri.parse('${AppConfig.baseUrl}/api/roles/$id'),
        headers: await sessionHeaders,
      ),
      'Gagal memuat role',
    );
    final data = payload['data'];
    if (data is! Map) {
      throw const RoleServiceException('Data role tidak valid');
    }
    return RoleModel.fromJson(Map<String, dynamic>.from(data));
  }

  Future<int> createRole(RoleModel role) async {
    final payload = await _requestObject(
      () async => client.post(
        Uri.parse('${AppConfig.baseUrl}/api/roles'),
        headers: await sessionJsonHeaders,
        body: await encryptJsonBody(role.toJson()),
      ),
      'Gagal membuat role',
    );
    return _parseId(payload['id']);
  }

  Future<void> updateRole(RoleModel role) async {
    await _requestObject(
      () async => client.put(
        Uri.parse('${AppConfig.baseUrl}/api/roles/${role.id}'),
        headers: await sessionJsonHeaders,
        body: await encryptJsonBody(role.toJson()),
      ),
      'Gagal memperbarui role',
    );
  }

  Future<void> deleteRole(int id) async {
    await _requestObject(
      () async => client.delete(
        Uri.parse('${AppConfig.baseUrl}/api/roles/$id'),
        headers: await sessionHeaders,
      ),
      'Gagal menghapus role',
    );
  }

  Future<List<RoleModel>> _getRoleList(Uri uri) async {
    try {
      final response = await client.get(uri, headers: await sessionHeaders);
      final decoded = jsonDecode(response.body);
      if (decoded is! Map) {
        throw const RoleServiceException('Format respons role tidak valid');
      }
      final payload = Map<String, dynamic>.from(decoded);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw RoleServiceException(
          payload['message']?.toString() ??
              payload['error']?.toString() ??
              'Gagal memuat role (${response.statusCode})',
        );
      }
      final data = payload['data'];
      if (data is! List) {
        throw const RoleServiceException('Data role tidak valid');
      }
      return data
          .whereType<Map>()
          .map((item) => RoleModel.fromJson(Map<String, dynamic>.from(item)))
          .toList(growable: false);
    } on RoleServiceException {
      rethrow;
    } on Exception catch (error) {
      throw RoleServiceException('Tidak dapat memuat role: $error');
    }
  }

  Future<Map<String, dynamic>> _requestObject(
    Future<http.Response> Function() request,
    String fallback,
  ) async {
    try {
      final response = await request();
      final decoded = jsonDecode(response.body);
      if (decoded is! Map) {
        throw const RoleServiceException('Format respons role tidak valid');
      }
      final payload = Map<String, dynamic>.from(decoded);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw RoleServiceException(
          payload['message']?.toString() ??
              payload['error']?.toString() ??
              '$fallback (${response.statusCode})',
        );
      }
      return payload;
    } on RoleServiceException {
      rethrow;
    } on Exception catch (error) {
      throw RoleServiceException('$fallback: $error');
    }
  }

  int _parseId(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    final id = int.tryParse(value?.toString() ?? '');
    if (id != null) return id;
    throw const RoleServiceException('Server tidak mengembalikan ID role');
  }
}

class RoleActionModel {
  const RoleActionModel({
    required this.value,
    required this.label,
    required this.sortOrder,
    required this.active,
  });

  final String value;
  final String label;
  final int sortOrder;
  final bool active;

  factory RoleActionModel.fromJson(Map<String, dynamic> json) =>
      RoleActionModel(
        value: json['value']?.toString() ?? '',
        label: json['label']?.toString() ?? '',
        sortOrder: int.tryParse(json['sort_order']?.toString() ?? '') ?? 0,
        active: json['active'] == true,
      );
}

class RoleServiceException implements Exception {
  const RoleServiceException(this.message);

  final String message;

  @override
  String toString() => message;
}
