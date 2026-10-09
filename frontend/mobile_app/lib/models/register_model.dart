class RegisterModel {
  final String name;
  final String email;
  final String phone;
  final String password;

  const RegisterModel({
    required this.name,
    required this.email,
    required this.phone,
    required this.password,
  });

  Map<String, dynamic> toJson() {
    return {'name': name, 'email': email, 'phone': phone, 'password': password};
  }
}

class RegisterResponse {
  final bool success;
  final String message;
  final String? customerId;

  const RegisterResponse({
    required this.success,
    required this.message,
    this.customerId,
  });

  factory RegisterResponse.fromJson(Map<String, dynamic> json) {
    return RegisterResponse(
      success: json['success'] ?? false,
      message: json['message'] ?? '',
      customerId: json['customerId']?.toString(),
    );
  }
}
