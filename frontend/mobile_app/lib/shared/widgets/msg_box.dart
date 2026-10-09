import 'package:flutter/material.dart';

class MsgBox extends StatelessWidget {
  const MsgBox({
    super.key,
    required this.message,
    this.title = 'Informasi',
    this.buttonText = 'OK',
  });

  final String message;
  final String title;
  final String buttonText;

  static Future<void> show(
    BuildContext context, {
    required String message,
    String title = 'Informasi',
    String buttonText = 'OK',
  }) {
    return showDialog<void>(
      context: context,
      builder: (_) =>
          MsgBox(title: title, message: message, buttonText: buttonText),
    );
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(title),
      content: Text(message),
      actions: [
        FilledButton(
          onPressed: () => Navigator.of(context).pop(),
          child: Text(buttonText),
        ),
      ],
    );
  }
}
