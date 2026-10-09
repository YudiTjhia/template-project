import 'dart:convert';
import 'dart:math';
import 'dart:typed_data';

import 'package:mobile_app/config/app_config.dart';
import 'package:http/http.dart' as http;
import 'package:pointycastle/asn1.dart';
import 'package:pointycastle/export.dart';

const encryptedRequestContentType = 'application/vnd.pos.encrypted+json';

class RequestEncryption {
  static Future<_ServerKey>? _serverKeyFuture;

  static void resetServerKeyCache() {
    _serverKeyFuture = null;
  }

  static Future<String> encrypt(http.Client client, String plaintext) async {
    final serverKey = await (_serverKeyFuture ??= _loadServerKey(client));
    final aesKey = _randomBytes(32);
    final nonce = _randomBytes(12);

    final rsa = OAEPEncoding.withSHA256(RSAEngine())
      ..init(true, PublicKeyParameter<RSAPublicKey>(serverKey.publicKey));
    final encryptedKey = rsa.process(aesKey);

    final aes = GCMBlockCipher(AESEngine())
      ..init(
        true,
        AEADParameters(KeyParameter(aesKey), 128, nonce, Uint8List(0)),
      );
    final ciphertext = aes.process(Uint8List.fromList(utf8.encode(plaintext)));

    return jsonEncode({
      'keyId': serverKey.keyId,
      'encryptedKey': base64Encode(encryptedKey),
      'nonce': base64Encode(nonce),
      'ciphertext': base64Encode(ciphertext),
    });
  }

  static Future<_ServerKey> _loadServerKey(http.Client client) async {
    final response = await client.get(
      Uri.parse('${AppConfig.baseUrl}/.well-known/request-encryption-key'),
      headers: {
        'Accept': 'application/json',
        'X-Client-ID': await AppConfig.clientId,
        'X-App-Source': AppConfig.appSource,
      },
    );
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw StateError(
        'Gagal mengambil kunci enkripsi (${response.statusCode})',
      );
    }

    final payload = jsonDecode(response.body);
    if (payload is! Map<String, dynamic>) {
      throw const FormatException('Format kunci enkripsi tidak valid');
    }
    final keyId = payload['keyId']?.toString() ?? '';
    final pem = payload['publicKey']?.toString() ?? '';
    if (keyId.isEmpty || pem.isEmpty) {
      throw const FormatException('Kunci publik backend tidak tersedia');
    }
    return _ServerKey(keyId, _parseRsaPublicKey(pem));
  }

  static RSAPublicKey _parseRsaPublicKey(String pem) {
    final encoded = pem
        .replaceAll('-----BEGIN PUBLIC KEY-----', '')
        .replaceAll('-----END PUBLIC KEY-----', '')
        .replaceAll(RegExp(r'\s'), '');
    final subjectPublicKeyInfo =
        ASN1Parser(Uint8List.fromList(base64Decode(encoded))).nextObject()
            as ASN1Sequence;
    final publicKeyBits = subjectPublicKeyInfo.elements![1] as ASN1BitString;
    final rsaPublicKey =
        ASN1Parser(Uint8List.fromList(publicKeyBits.stringValues!)).nextObject()
            as ASN1Sequence;
    final modulus = (rsaPublicKey.elements![0] as ASN1Integer).integer!;
    final exponent = (rsaPublicKey.elements![1] as ASN1Integer).integer!;
    return RSAPublicKey(modulus, exponent);
  }

  static Uint8List _randomBytes(int length) {
    final random = Random.secure();
    return Uint8List.fromList(
      List<int>.generate(length, (_) => random.nextInt(256)),
    );
  }
}

class _ServerKey {
  const _ServerKey(this.keyId, this.publicKey);

  final String keyId;
  final RSAPublicKey publicKey;
}
