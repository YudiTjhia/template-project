import 'package:mobile_app/models/role_model.dart';
import 'package:mobile_app/services/role_service.dart';

/// In-memory role data for UI development and demos.
/// Set `USE_DUMMY_ROLES=false` to use the backend [RoleService].
class DummyRoleService extends RoleService {
  DummyRoleService();

  static final List<RoleModel> _roles = [
    const RoleModel(
      id: 1,
      name: 'SUPER_ADMIN',
      actions: [
        'view_pos_dashboard',
        'view_sales_order',
        'create_sales_order',
        'update_sales_order',
        'delete_sales_order',
        'view_item',
        'create_item',
        'update_item',
        'delete_item',
      ],
    ),
    const RoleModel(
      id: 2,
      name: 'MANAGER',
      actions: [
        'view_pos_dashboard',
        'view_sales_order',
        'create_sales_order',
        'update_sales_order',
        'view_item',
        'select_location',
      ],
    ),
    const RoleModel(
      id: 3,
      name: 'CASHIER',
      actions: [
        'view_pos_dashboard',
        'view_sales_order',
        'create_sales_order',
        'checkout_sales_order',
        'cancel_sales_order',
      ],
    ),
  ];

  static const _actions = <RoleActionModel>[
    RoleActionModel(
      value: 'view_pos_dashboard',
      label: 'View POS Dashboard',
      sortOrder: 10,
      active: true,
    ),
    RoleActionModel(
      value: 'view_sales_order',
      label: 'View Sales Order',
      sortOrder: 100,
      active: true,
    ),
    RoleActionModel(
      value: 'create_sales_order',
      label: 'Create Sales Order',
      sortOrder: 110,
      active: true,
    ),
    RoleActionModel(
      value: 'update_sales_order',
      label: 'Update Sales Order',
      sortOrder: 120,
      active: true,
    ),
    RoleActionModel(
      value: 'delete_sales_order',
      label: 'Delete Sales Order',
      sortOrder: 130,
      active: true,
    ),
    RoleActionModel(
      value: 'checkout_sales_order',
      label: 'Checkout Sales Order',
      sortOrder: 140,
      active: true,
    ),
    RoleActionModel(
      value: 'cancel_sales_order',
      label: 'Cancel Sales Order',
      sortOrder: 150,
      active: true,
    ),
    RoleActionModel(
      value: 'view_item',
      label: 'View Item',
      sortOrder: 400,
      active: true,
    ),
    RoleActionModel(
      value: 'create_item',
      label: 'Create Item',
      sortOrder: 410,
      active: true,
    ),
    RoleActionModel(
      value: 'update_item',
      label: 'Update Item',
      sortOrder: 420,
      active: true,
    ),
    RoleActionModel(
      value: 'delete_item',
      label: 'Delete Item',
      sortOrder: 430,
      active: true,
    ),
    RoleActionModel(
      value: 'select_location',
      label: 'Select Active Location',
      sortOrder: 940,
      active: true,
    ),
  ];

  @override
  Future<List<RoleModel>> getRoles() async => List.unmodifiable(_roles);

  @override
  Future<List<RoleActionModel>> getRoleActions() async => _actions;

  @override
  Future<RoleModel> getRole(int id) async {
    return _roles.firstWhere(
      (role) => role.id == id,
      orElse: () =>
          throw const RoleServiceException('Role dummy tidak ditemukan'),
    );
  }

  @override
  Future<int> createRole(RoleModel role) async {
    final id =
        _roles.fold<int>(
          0,
          (largest, item) => item.id > largest ? item.id : largest,
        ) +
        1;
    _roles.add(
      RoleModel(id: id, name: role.name, actions: List.of(role.actions)),
    );
    return id;
  }

  @override
  Future<void> updateRole(RoleModel role) async {
    final index = _roles.indexWhere((item) => item.id == role.id);
    if (index < 0) {
      throw const RoleServiceException('Role dummy tidak ditemukan');
    }
    _roles[index] = RoleModel(
      id: role.id,
      name: role.name,
      actions: List.of(role.actions),
    );
  }

  @override
  Future<void> deleteRole(int id) async {
    final lengthBefore = _roles.length;
    _roles.removeWhere((role) => role.id == id);
    if (_roles.length == lengthBefore) {
      throw const RoleServiceException('Role dummy tidak ditemukan');
    }
  }
}
