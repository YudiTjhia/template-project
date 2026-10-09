import 'package:flutter/material.dart';

class ConfirmBox extends StatelessWidget {
  const ConfirmBox({
    super.key,
    required this.message,
    this.title = 'Konfirmasi',
    this.confirmText = 'Ya',
    this.cancelText = 'Batal',
  });

  final String message;
  final String title;
  final String confirmText;
  final String cancelText;

  static Future<bool> show(
    BuildContext context, {
    required String message,
    String title = 'Konfirmasi',
    String confirmText = 'Ya',
    String cancelText = 'Batal',
  }) async {
    return await showDialog<bool>(
          context: context,
          builder: (_) => ConfirmBox(
            title: title,
            message: message,
            confirmText: confirmText,
            cancelText: cancelText,
          ),
        ) ??
        false;
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(title),
      content: Text(message),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(false),
          child: Text(cancelText),
        ),
        FilledButton(
          onPressed: () => Navigator.of(context).pop(true),
          child: Text(confirmText),
        ),
      ],
    );
  }
}
