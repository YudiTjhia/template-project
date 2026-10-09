class UserModel {
  const UserModel({
    required this.id,
    required this.username,
    required this.email,
    required this.mobilePhone,
    required this.userType,
    required this.loginType,
    required this.userRoles,
    required this.isActive,
  });

  final int id;
  final String username;
  final String email;
  final String mobilePhone;
  final String userType;
  final String loginType;
  final List<String> userRoles;
  final bool isActive;

  factory UserModel.fromJson(Map<String, dynamic> json) => UserModel(
    id: _asInt(json['id']),
    username: json['username']?.toString() ?? '',
    email: json['user_email']?.toString() ?? '',
    mobilePhone: json['mobilephone']?.toString() ?? '',
    userType: json['user_type']?.toString() ?? 'USER',
    loginType: json['login_type']?.toString() ?? 'cashier',
    userRoles: (json['user_roles'] as List? ?? const [])
        .map((role) => role.toString())
        .toList(growable: false),
    isActive: json['is_active'] == true,
  );

  Map<String, dynamic> toJson({String password = '', String pin = ''}) => {
    'username': username.trim(),
    'user_email': email.trim(),
    'mobilephone': mobilePhone.trim(),
    'user_type': userType,
    'login_type': loginType,
    'user_roles': userRoles,
    'is_active': isActive,
    if (password.isNotEmpty) 'user_password': password,
    if (pin.isNotEmpty) 'pin': pin,
  };

  static int _asInt(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    return int.tryParse(value?.toString() ?? '') ?? 0;
  }
}
