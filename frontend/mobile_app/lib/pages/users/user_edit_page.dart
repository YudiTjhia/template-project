import 'package:flutter/material.dart';

import 'package:mobile_app/models/role_model.dart';
import 'package:mobile_app/services/role_service.dart';

import '../../shared/widgets/edit_component.dart';
import '../../shared/widgets/edit_fields.dart';

import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/user_service.dart';

class UserEditPage extends StatefulWidget {
  const UserEditPage({super.key, this.userId, this.service, this.roleService});

  final int? userId;
  final UserService? service;
  final RoleService? roleService;

  @override
  State<UserEditPage> createState() => _UserEditPageState();
}

class _UserEditPageState extends State<UserEditPage> {
  late final UserService _service = widget.service ?? UserService();
  late final RoleService _roleService = widget.roleService ?? RoleService();
  final _username = TextEditingController();
  final _email = TextEditingController();
  final _mobilePhone = TextEditingController();
  final _password = TextEditingController();
  List<RoleModel> _roles = const [];
  Set<String> _selectedRoles = {};
  String _userType = 'USER';
  String? _originalUserType;
  String _loginType = 'cashier';
  bool _isActive = true;
  bool _loading = true;
  bool _saving = false;
  String? _error;

  bool get _isCreate => widget.userId == null;

  @override
  void initState() {
    super.initState();
    _loadForm();
  }

  @override
  void dispose() {
    _username.dispose();
    _email.dispose();
    _mobilePhone.dispose();
    _password.dispose();
    super.dispose();
  }

  Future<void> _loadForm() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final rolesFuture = _roleService.getRoles();
      final userFuture = widget.userId == null
          ? Future<UserModel?>.value(null)
          : _service.getUser(widget.userId!);
      final results = await Future.wait<Object?>([rolesFuture, userFuture]);
      final roles = results[0] as List<RoleModel>;
      final user = results[1] as UserModel?;
      if (!mounted) return;
      setState(() {
        _roles = roles;
        if (user != null) {
          _username.text = user.username;
          _email.text = user.email;
          _mobilePhone.text = user.mobilePhone;
          _userType = user.userType;
          _originalUserType = user.userType;
          _loginType = user.loginType;
          _selectedRoles = user.userRoles.toSet();
          _isActive = user.isActive;
        }
        _loading = false;
      });
    } on UserServiceException catch (error) {
      _setLoadError(error.message);
    } on RoleServiceException catch (error) {
      _setLoadError(error.message);
    }
  }

  void _setLoadError(String message) {
    if (!mounted) return;
    setState(() {
      _error = message;
      _loading = false;
    });
  }

  List<EditFieldMetadata> get _fields => [
    EditFieldMetadata(
      name: 'username',
      label: 'Username',
      controller: _username,
      required: true,
    ),
    EditFieldMetadata(
      name: 'user_email',
      label: 'Email',
      type: EditFieldType.email,
      controller: _email,
      required: true,
      validator: (value) {
        final email = value?.trim() ?? '';
        if (!RegExp(r'^[^\s@]+@[^\s@]+\.[^\s@]+$').hasMatch(email)) {
          return 'Format email tidak valid';
        }
        return null;
      },
    ),
    EditFieldMetadata(
      name: 'mobilephone',
      label: 'Nomor telepon',
      type: EditFieldType.number,
      controller: _mobilePhone,
      required: true,
    ),
    EditFieldMetadata(
      name: 'user_type',
      label: 'Tipe user',
      type: EditFieldType.dropdown,
      initialValue: _userType,
      options: const [
        EditFieldOption(value: 'USER', label: 'User'),
        EditFieldOption(value: 'SUPER_ADMIN', label: 'Super admin'),
      ],
      onChanged: (value) {
        if (value != null) setState(() => _userType = value);
      },
    ),
    EditFieldMetadata(
      name: 'login_type',
      label: 'Tipe login',
      type: EditFieldType.dropdown,
      initialValue: _loginType,
      options: const [
        EditFieldOption(value: 'cashier', label: 'Kasir'),
        EditFieldOption(value: 'customer_table', label: 'Meja pelanggan'),
      ],
      onChanged: (value) {
        if (value != null) setState(() => _loginType = value);
      },
    ),
    EditFieldMetadata(
      name: 'user_roles',
      label: 'Role',
      type: EditFieldType.multiSelect,
      initialValues: _selectedRoles.toList(),
      options: _roles
          .map((role) => EditFieldOption(value: role.name, label: role.name))
          .toList(growable: false),
      required: _userType == 'USER',
      helperText: _roles.isEmpty ? 'Belum ada role yang dibuat' : null,
      onMultiChanged: (values) =>
          setState(() => _selectedRoles = values.toSet()),
    ),
    EditFieldMetadata(
      name: 'is_active',
      label: 'Status user',
      type: EditFieldType.dropdown,
      initialValue: _isActive ? 'active' : 'inactive',
      options: const [
        EditFieldOption(value: 'active', label: 'Aktif'),
        EditFieldOption(value: 'inactive', label: 'Nonaktif'),
      ],
      onChanged: (value) {
        if (value != null) setState(() => _isActive = value == 'active');
      },
    ),
    EditFieldMetadata(
      name: 'user_password',
      label: _isCreate ? 'Password' : 'Password baru (opsional)',
      type: EditFieldType.password,
      controller: _password,
      required: _isCreate,
      validator: (value) {
        if ((value ?? '').isNotEmpty && value!.length < 8) {
          return 'Password minimal 8 karakter';
        }
        return null;
      },
    ),
  ];

  UserModel _formUser() => UserModel(
    id: widget.userId ?? 0,
    username: _username.text,
    email: _email.text,
    mobilePhone: _mobilePhone.text,
    userType: _userType,
    loginType: _loginType,
    userRoles: _userType == 'USER' ? _selectedRoles.toList() : const [],
    isActive: _isActive,
  );

  Future<void> _save() async {
    if (_saving) return;
    setState(() => _saving = true);
    try {
      if (_isCreate) {
        await _service.createUser(_formUser(), password: _password.text);
      } else {
        await _service.updateUser(_formUser(), password: _password.text);
      }
      if (!mounted) return;
      Navigator.of(context).pop(true);
    } on UserServiceException catch (error) {
      if (!mounted) return;
      setState(() => _saving = false);
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(error.message)));
    }
  }

  Future<void> _delete() async {
    final shouldDelete = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Hapus user?'),
        content: const Text('User ini akan dihapus dari akun.'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Batal'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Hapus'),
          ),
        ],
      ),
    );
    if (shouldDelete != true || !mounted) return;
    setState(() => _saving = true);
    try {
      await _service.deleteUser(widget.userId!);
      if (mounted) Navigator.of(context).pop(true);
    } on UserServiceException catch (error) {
      if (!mounted) return;
      setState(() => _saving = false);
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(error.message)));
    }
  }

  @override
  Widget build(BuildContext context) {
    final title = _isCreate ? 'Tambah User' : 'Edit User';
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(_error!, textAlign: TextAlign.center),
                    const SizedBox(height: 12),
                    OutlinedButton.icon(
                      onPressed: _loadForm,
                      icon: const Icon(Icons.refresh),
                      label: const Text('Coba lagi'),
                    ),
                  ],
                ),
              ),
            )
          : SafeArea(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(20),
                child: EditComponent(
                  fields: _fields,
                  mode: _isCreate ? EditFormMode.create : EditFormMode.edit,
                  isLoading: _saving,
                  onCancel: () => Navigator.of(context).pop(false),
                  onSaveNew: _save,
                  onUpdate: _save,
                  onDelete: _isCreate || _originalUserType == 'SUPER_ADMIN'
                      ? null
                      : _delete,
                ),
              ),
            ),
    );
  }
}
