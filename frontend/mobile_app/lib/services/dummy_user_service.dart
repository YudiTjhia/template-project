import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/user_service.dart';

/// Demo list source. Replace it with [UserService] by setting
/// `USE_DUMMY_USERS=false` in the app's `.env` file.
class DummyUserService extends UserService {
  DummyUserService();

  static const _users = <UserModel>[
    UserModel(
      id: 1,
      username: 'admin.demo',
      email: 'admin@example.com',
      mobilePhone: '081234567890',
      userType: 'SUPER_ADMIN',
      loginType: 'cashier',
      userRoles: ['SUPER_ADMIN'],
      isActive: true,
    ),
    UserModel(
      id: 2,
      username: 'store.manager',
      email: 'manager@example.com',
      mobilePhone: '081234567891',
      userType: 'USER',
      loginType: 'cashier',
      userRoles: ['MANAGER'],
      isActive: true,
    ),
    UserModel(
      id: 3,
      username: 'cashier.demo',
      email: 'cashier@example.com',
      mobilePhone: '081234567892',
      userType: 'USER',
      loginType: 'cashier',
      userRoles: ['CASHIER'],
      isActive: true,
    ),
    UserModel(
      id: 4,
      username: 'inactive.user',
      email: 'inactive@example.com',
      mobilePhone: '081234567893',
      userType: 'USER',
      loginType: 'cashier',
      userRoles: ['CASHIER'],
      isActive: false,
    ),
  ];

  @override
  Future<List<UserModel>> getUsers({String search = ''}) async {
    final query = search.trim().toLowerCase();
    if (query.isEmpty) return _users;
    return _users
        .where(
          (user) =>
              user.username.toLowerCase().contains(query) ||
              user.email.toLowerCase().contains(query) ||
              user.mobilePhone.toLowerCase().contains(query) ||
              user.userRoles.any((role) => role.toLowerCase().contains(query)),
        )
        .toList(growable: false);
  }

  @override
  Future<UserModel> getUser(int id) async {
    return _users.firstWhere(
      (user) => user.id == id,
      orElse: () =>
          throw const UserServiceException('User dummy tidak ditemukan'),
    );
  }
}
