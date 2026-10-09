import 'package:mobile_app/config/app_config.dart';

import '../models/register_model.dart';

import 'package:mobile_app/shared/services/base_service.dart';
import 'package:mobile_app/shared/services/api_helper.dart';

class RegisterService extends BaseService {
  RegisterService({super.client});

  Future<RegisterResponse> register(RegisterModel customer) async {
    final uri = Uri.parse('${AppConfig.baseUrl}/api/customers/register');

    try {
      final response = await client.post(
        uri,
        headers: await jsonHeaders,
        body: await encryptJsonBody(customer.toJson()),
      );
      final data = ApiHelper.decodeApiObject(response.body);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        return RegisterResponse(
          success: false,
          message:
              data['message']?.toString() ??
              data['error']?.toString() ??
              'Register gagal (${response.statusCode})',
        );
      }

      final result = RegisterResponse.fromJson(data);
      if (!result.success) {
        return RegisterResponse(
          success: false,
          message: result.message.isNotEmpty
              ? result.message
              : 'Server tidak mengonfirmasi registrasi customer',
        );
      }
      return result;
    } on ApiException catch (error) {
      return RegisterResponse(success: false, message: error.message);
    } on FormatException {
      return const RegisterResponse(
        success: false,
        message: 'Format respons server tidak valid',
      );
    } on Exception catch (error) {
      return RegisterResponse(
        success: false,
        message: 'Tidak dapat terhubung ke server: $error',
      );
    }
  }
}
