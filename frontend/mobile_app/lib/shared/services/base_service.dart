import 'dart:convert';

import 'package:mobile_app/shared/config/app_config.dart';
import 'package:http/http.dart' as http;

import '../helpers/session_helper.dart';

import 'request_encryption.dart';

abstract class BaseService {
  BaseService({http.Client? client}) : client = client ?? http.Client();

  final http.Client client;

  Future<Map<String, String>> get headers async => {
    'Accept': 'application/json',
    'X-Client-ID': await AppConfig.clientId,
    'X-App-Source': AppConfig.appSource,
  };

  Future<Map<String, String>> get jsonHeaders async => {
    ...await headers,
    'Content-Type': encryptedRequestContentType,
  };

  Future<String> encryptJsonBody(Object payload) =>
      RequestEncryption.encrypt(client, jsonEncode(payload));

  Future<Map<String, String>> get sessionHeaders async {
    final token = await SessionHelper.getToken();
    return {
      ...await headers,
      if (token != null && token.isNotEmpty) ...{
        'X-Session-Token': token,
        'Authorization': 'Bearer $token',
      },
    };
  }

  Future<Map<String, String>> get sessionJsonHeaders async => {
    ...(await sessionHeaders),
    'Content-Type': encryptedRequestContentType,
  };
}
