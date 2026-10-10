import 'package:flutter/material.dart';
import 'package:mobile_app/models/role_model.dart';
import 'package:mobile_app/services/role_service.dart';
import 'package:mobile_app/shared/widgets/edit_component.dart';
import 'package:mobile_app/shared/widgets/edit_fields.dart';
import 'package:mobile_app/navigation/app_drawer_menu.dart';

import 'role_actions_lookup_dialog.dart';
import '../../shared/widgets/responsive_page_layout.dart';

class RoleEditPage extends StatefulWidget {
  const RoleEditPage({
    super.key,
    this.roleId,
    this.service,
    this.enableDrawer = true,
  });

  final int? roleId;
  final RoleService? service;
  final bool enableDrawer;

  @override
  State<RoleEditPage> createState() => _RoleEditPageState();
}

class _RoleEditPageState extends State<RoleEditPage> {
  late final RoleService _service = widget.service ?? RoleService();
  final _nameController = TextEditingController();
  Set<String> _selectedActions = {};
  bool _loading = true;
  bool _saving = false;
  String? _error;

  bool get _isCreate => widget.roleId == null;

  @override
  void initState() {
    super.initState();
    _loadForm();
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _loadForm() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final roleFuture = widget.roleId == null
          ? Future<RoleModel?>.value(null)
          : _service.getRole(widget.roleId!);
      final role = await roleFuture;
      if (!mounted) return;
      setState(() {
        if (role != null) {
          _nameController.text = role.name;
          _selectedActions = role.actions.toSet();
        }
        _loading = false;
      });
    } on RoleServiceException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _loading = false;
      });
    }
  }

  List<EditFieldMetadata> get _fields => [
    EditFieldMetadata(
      name: 'role_name',
      label: 'Nama role',
      controller: _nameController,
      required: true,
    ),
  ];

  Future<void> _openActionLookup() async {
    final actions = await RoleActionsLookupDialog.show(
      context,
      service: _service,
      selectedValues: _selectedActions.toList(),
    );
    if (actions == null || !mounted) return;
    setState(() {
      _selectedActions = actions.map((action) => action.value).toSet();
    });
  }

  Future<void> _save() async {
    if (_saving) return;
    setState(() => _saving = true);
    final role = RoleModel(
      id: widget.roleId ?? 0,
      name: _nameController.text.trim(),
      actions: _selectedActions.toList()..sort(),
    );
    try {
      if (_isCreate) {
        await _service.createRole(role);
      } else {
        await _service.updateRole(role);
      }
      if (mounted) Navigator.of(context).pop(true);
    } on RoleServiceException catch (error) {
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
        title: const Text('Hapus role?'),
        content: const Text('Role ini akan dihapus dari akun.'),
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
      await _service.deleteRole(widget.roleId!);
      if (mounted) Navigator.of(context).pop(true);
    } on RoleServiceException catch (error) {
      if (!mounted) return;
      setState(() => _saving = false);
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(error.message)));
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(_isCreate ? 'Tambah Role' : 'Edit Role')),
      drawer: widget.enableDrawer
          ? const AppDrawerMenu(selectedRoute: '/roles')
          : null,
      body: ResponsivePageLayout(
        maxWidth: 760,
        safeArea: true,
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
            ? Center(
                child: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 24),
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
                  padding: const EdgeInsets.symmetric(vertical: 20),
                  child: EditComponent(
                    fields: _fields,
                    additionalFields: [_buildActionSelector()],
                    mode: _isCreate ? EditFormMode.create : EditFormMode.edit,
                    isLoading: _saving,
                    onCancel: () => Navigator.of(context).pop(false),
                    onSaveNew: _save,
                    onUpdate: _save,
                    onDelete: _isCreate ? null : _delete,
                  ),
                ),
              ),
      ),
    );
  }

  Widget _buildActionSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Hak akses', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          onPressed: _saving ? null : _openActionLookup,
          icon: const Icon(Icons.tune),
          label: const Text('Pilih hak akses'),
        ),
        if (_selectedActions.isEmpty)
          const Padding(
            padding: EdgeInsets.only(top: 8),
            child: Text('Belum ada action dipilih.'),
          )
        else
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Wrap(
              spacing: 8,
              runSpacing: 4,
              children: [
                for (final action in _selectedActions)
                  Chip(
                    label: Text(action),
                    onDeleted: _saving
                        ? null
                        : () => setState(() => _selectedActions.remove(action)),
                  ),
              ],
            ),
          ),
      ],
    );
  }
}
