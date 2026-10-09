import 'package:flutter/material.dart';
import 'package:mobile_app/services/role_service.dart';
import 'package:mobile_app/shared/widgets/lookup_dialog.dart';

/// Loads the available role actions and returns the selected action codes.
class RoleActionsLookupDialog {
  const RoleActionsLookupDialog._();

  static Future<List<RoleActionModel>?> show(
    BuildContext context, {
    required RoleService service,
    List<String> selectedValues = const [],
  }) async {
    List<RoleActionModel> actions;
    try {
      actions = await service.getRoleActions();
    } on RoleServiceException catch (error) {
      if (!context.mounted) return null;
      await showDialog<void>(
        context: context,
        builder: (dialogContext) => AlertDialog(
          title: const Text('Gagal memuat hak akses'),
          content: Text(error.message),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: const Text('Tutup'),
            ),
          ],
        ),
      );
      return null;
    }

    if (!context.mounted) return null;
    final selectedItems = actions
        .where((action) => selectedValues.contains(action.value))
        .toList(growable: false);
    final result = await LookupDialog.show<RoleActionModel>(
      context,
      items: actions,
      title: 'Pilih hak akses',
      labelBuilder: (action) => action.label,
      subtitleBuilder: (action) => action.value,
      selectedItems: selectedItems,
      multiSelect: true,
      searchHint: 'Cari kode atau nama action',
      emptyMessage: 'Action tidak ditemukan',
    );
    if (result is List<RoleActionModel>) return result;
    return null;
  }
}
