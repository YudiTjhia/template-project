import 'package:flutter/material.dart';

import 'action_buttons.dart';
import 'edit_fields.dart';

enum EditFormMode { create, edit, view }

/// Form layout that combines metadata-driven fields with standard actions.
class EditComponent extends StatefulWidget {
  const EditComponent({
    super.key,
    required this.fields,
    required this.mode,
    this.formKey,
    this.additionalFields = const [],
    this.onCancel,
    this.onSaveNew,
    this.onUpdate,
    this.onDelete,
    this.showCancel = true,
    this.showSaveNew,
    this.showUpdate,
    this.showDelete,
    this.enableCancel = true,
    this.enableSaveNew = true,
    this.enableUpdate = true,
    this.enableDelete = true,
    this.isLoading = false,
    this.fieldSpacing = 16,
    this.actionSpacing = 8,
    this.padding = EdgeInsets.zero,
  });

  final List<EditFieldMetadata> fields;
  final EditFormMode mode;
  final GlobalKey<FormState>? formKey;
  final List<Widget> additionalFields;
  final VoidCallback? onCancel;
  final VoidCallback? onSaveNew;
  final VoidCallback? onUpdate;
  final VoidCallback? onDelete;
  final bool showCancel;
  final bool? showSaveNew;
  final bool? showUpdate;
  final bool? showDelete;
  final bool enableCancel;
  final bool enableSaveNew;
  final bool enableUpdate;
  final bool enableDelete;
  final bool isLoading;
  final double fieldSpacing;
  final double actionSpacing;
  final EdgeInsetsGeometry padding;

  @override
  State<EditComponent> createState() => _EditComponentState();
}

class _EditComponentState extends State<EditComponent> {
  late final GlobalKey<FormState> _formKey =
      widget.formKey ?? GlobalKey<FormState>();

  bool _validate() => _formKey.currentState?.validate() ?? false;

  List<EditFieldMetadata> get _fields => widget.mode == EditFormMode.view
      ? widget.fields.map((field) => field.copyWith(readOnly: true)).toList()
      : widget.fields;

  @override
  Widget build(BuildContext context) {
    final isCreate = widget.mode == EditFormMode.create;
    final isEdit = widget.mode == EditFormMode.edit;

    return Padding(
      padding: widget.padding,
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            EditFields(fields: _fields, spacing: widget.fieldSpacing),
            if (widget.additionalFields.isNotEmpty) ...[
              SizedBox(height: widget.fieldSpacing),
              ...widget.additionalFields,
            ],
            const SizedBox(height: 24),
            ActionButtons(
              onCancel: widget.onCancel,
              onSaveNew: _validateAndSave,
              onUpdate: _validateAndUpdate,
              onDelete: widget.onDelete,
              showCancel: widget.showCancel,
              showSaveNew: widget.showSaveNew ?? isCreate,
              showUpdate: widget.showUpdate ?? isEdit,
              showDelete: widget.showDelete ?? isEdit,
              enableCancel: widget.enableCancel,
              enableSaveNew: widget.enableSaveNew && widget.onSaveNew != null,
              enableUpdate: widget.enableUpdate && widget.onUpdate != null,
              enableDelete: widget.enableDelete,
              isLoading: widget.isLoading,
              spacing: widget.actionSpacing,
            ),
          ],
        ),
      ),
    );
  }

  void _validateAndSave() {
    if (_validate()) widget.onSaveNew?.call();
  }

  void _validateAndUpdate() {
    if (_validate()) widget.onUpdate?.call();
  }
}
