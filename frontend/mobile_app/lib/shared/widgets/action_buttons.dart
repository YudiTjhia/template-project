import 'package:flutter/material.dart';

/// Standard cancel, create, update, and delete actions.
class ActionButtons extends StatelessWidget {
  const ActionButtons({
    super.key,
    this.onCancel,
    this.onSaveNew,
    this.onUpdate,
    this.onDelete,
    this.showCancel = true,
    this.showSaveNew = true,
    this.showUpdate = true,
    this.showDelete = true,
    this.enableCancel = true,
    this.enableSaveNew = true,
    this.enableUpdate = true,
    this.enableDelete = true,
    this.isLoading = false,
    this.alignment = MainAxisAlignment.end,
    this.spacing = 8,
  });

  final VoidCallback? onCancel;
  final VoidCallback? onSaveNew;
  final VoidCallback? onUpdate;
  final VoidCallback? onDelete;
  final bool showCancel;
  final bool showSaveNew;
  final bool showUpdate;
  final bool showDelete;
  final bool enableCancel;
  final bool enableSaveNew;
  final bool enableUpdate;
  final bool enableDelete;
  final bool isLoading;
  final MainAxisAlignment alignment;
  final double spacing;

  @override
  Widget build(BuildContext context) {
    final buttons = <Widget>[];
    if (showCancel) {
      buttons.add(
        OutlinedButton.icon(
          onPressed: enableCancel && !isLoading ? onCancel : null,
          icon: const Icon(Icons.close),
          label: const Text('Batal'),
        ),
      );
    }
    if (showDelete) {
      buttons.add(
        OutlinedButton.icon(
          onPressed: enableDelete && !isLoading ? onDelete : null,
          icon: const Icon(Icons.delete_outline),
          label: const Text('Hapus'),
          style: OutlinedButton.styleFrom(foregroundColor: Colors.red),
        ),
      );
    }
    if (showSaveNew) {
      buttons.add(
        FilledButton.icon(
          onPressed: enableSaveNew && !isLoading ? onSaveNew : null,
          icon: const Icon(Icons.add),
          label: const Text('Simpan Baru'),
        ),
      );
    }
    if (showUpdate) {
      buttons.add(
        FilledButton.icon(
          onPressed: enableUpdate && !isLoading ? onUpdate : null,
          icon: const Icon(Icons.save_outlined),
          label: const Text('Perbarui'),
        ),
      );
    }

    return Wrap(
      alignment: _wrapAlignment,
      spacing: spacing,
      runSpacing: spacing,
      children: buttons,
    );
  }

  WrapAlignment get _wrapAlignment => switch (alignment) {
    MainAxisAlignment.start => WrapAlignment.start,
    MainAxisAlignment.center => WrapAlignment.center,
    MainAxisAlignment.end => WrapAlignment.end,
    MainAxisAlignment.spaceAround => WrapAlignment.spaceAround,
    MainAxisAlignment.spaceBetween => WrapAlignment.spaceBetween,
    MainAxisAlignment.spaceEvenly => WrapAlignment.spaceEvenly,
  };
}
