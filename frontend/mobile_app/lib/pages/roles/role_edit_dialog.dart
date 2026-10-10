import 'package:flutter/material.dart';
import 'package:mobile_app/models/role_model.dart';
import 'package:mobile_app/services/role_service.dart';
import 'package:mobile_app/shared/widgets/edit_component.dart';
import 'package:mobile_app/shared/widgets/edit_dialog.dart';
import 'package:mobile_app/shared/widgets/edit_fields.dart';
import 'package:mobile_app/shared/widgets/chips_list_widget.dart';

import 'role_actions_lookup_dialog.dart';

enum RoleEditMode { create, edit, view }

/// Role editor displayed as a dialog for create, edit, and view flows.
class RoleEditDialog extends StatefulWidget {
  const RoleEditDialog({
    super.key,
    required this.mode,
    this.roleId,
    this.service,
  });

  final RoleEditMode mode;
  final int? roleId;
  final RoleService? service;

  static Future<bool?> show(
    BuildContext context, {
    required RoleEditMode mode,
    int? roleId,
    RoleService? service,
  }) {
    final title = switch (mode) {
      RoleEditMode.create => 'Tambah Role',
      RoleEditMode.edit => 'Edit Role',
      RoleEditMode.view => 'Detail Role',
    };
    return EditDialog.show<bool>(
      context,
      title: title,
      maxWidth: 560,
      child: RoleEditDialog(mode: mode, roleId: roleId, service: service),
    );
  }

  @override
  State<RoleEditDialog> createState() => _RoleEditDialogState();
}

class _RoleEditDialogState extends State<RoleEditDialog> {
  late final RoleService _service = widget.service ?? RoleService();
  final _nameController = TextEditingController();
  Set<String> _selectedActions = {};
  bool _loading = true;
  bool _saving = false;
  String? _error;

  bool get _isCreate => widget.mode == RoleEditMode.create;
  bool get _isView => widget.mode == RoleEditMode.view;

  @override
  void initState() {
    super.initState();
    _loadRole();
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  Future<void> _loadRole() async {
    if (widget.roleId == null) {
      setState(() => _loading = false);
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final role = await _service.getRole(widget.roleId!);
      if (!mounted) return;
      setState(() {
        _nameController.text = role.name;
        _selectedActions = role.actions.toSet();
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
    if (_loading) {
      return const SizedBox(
        height: 180,
        child: Center(child: CircularProgressIndicator()),
      );
    }
    if (_error != null) {
      return SizedBox(
        height: 180,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(_error!, textAlign: TextAlign.center),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              onPressed: _loadRole,
              icon: const Icon(Icons.refresh),
              label: const Text('Coba lagi'),
            ),
          ],
        ),
      );
    }

    return EditComponent(
      fields: _fields,
      additionalFields: [_buildActionSelector()],
      mode: _isView
          ? EditFormMode.view
          : _isCreate
          ? EditFormMode.create
          : EditFormMode.edit,
      isLoading: _saving,
      showCancel: !_isView,
      showSaveNew: _isCreate,
      showUpdate: widget.mode == RoleEditMode.edit,
      showDelete: widget.mode == RoleEditMode.edit,
      onCancel: () => Navigator.of(context).pop(false),
      onSaveNew: _save,
      onUpdate: _save,
      onDelete: widget.mode == RoleEditMode.edit ? _delete : null,
    );
  }

  Widget _buildActionSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Hak akses', style: Theme.of(context).textTheme.titleSmall),
        const SizedBox(height: 8),
        OutlinedButton.icon(
          onPressed: _saving || _isView ? null : _openActionLookup,
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
            child: ChipsListWidget<String>(
              items: _selectedActions.toList()..sort(),
              onDeleted: _saving || _isView
                  ? null
                  : (action) => setState(() => _selectedActions.remove(action)),
            ),
          ),
      ],
    );
  }
}
