class LoginModel {
  const LoginModel({required this.identifier, required this.password});

  final String identifier;
  final String password;

  Map<String, dynamic> toJson() => {
    'identifier': identifier.trim(),
    'password': password,
  };
}
