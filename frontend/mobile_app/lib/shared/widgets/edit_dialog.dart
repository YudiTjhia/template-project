import 'package:flutter/material.dart';

/// Presents an edit form inside a responsive dialog.
///
/// The [child] can be an [EditComponent] or any custom form widget. The dialog
/// itself only handles presentation; form actions decide when to close it by
/// calling `Navigator.pop`.
class EditDialog extends StatelessWidget {
  const EditDialog({
    super.key,
    required this.title,
    required this.child,
    this.subtitle,
    this.maxWidth = 640,
    this.maxHeightFactor = 0.8,
    this.showCloseButton = true,
    this.barrierDismissible = true,
  });

  final String title;
  final String? subtitle;
  final Widget child;
  final double maxWidth;
  final double maxHeightFactor;
  final bool showCloseButton;
  final bool barrierDismissible;

  /// Opens an edit dialog and returns the value passed to `Navigator.pop`.
  static Future<T?> show<T>(
    BuildContext context, {
    required String title,
    required Widget child,
    String? subtitle,
    double maxWidth = 640,
    double maxHeightFactor = 0.8,
    bool showCloseButton = true,
    bool barrierDismissible = true,
  }) {
    return showDialog<T>(
      context: context,
      barrierDismissible: barrierDismissible,
      builder: (_) => EditDialog(
        title: title,
        subtitle: subtitle,
        maxWidth: maxWidth,
        maxHeightFactor: maxHeightFactor,
        showCloseButton: showCloseButton,
        barrierDismissible: barrierDismissible,
        child: child,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final maxHeight = MediaQuery.sizeOf(context).height * maxHeightFactor;

    return AlertDialog(
      title: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(title),
                if (subtitle != null) ...[
                  const SizedBox(height: 4),
                  Text(subtitle!, style: Theme.of(context).textTheme.bodySmall),
                ],
              ],
            ),
          ),
          if (showCloseButton)
            IconButton(
              tooltip: 'Tutup',
              onPressed: () => Navigator.of(context).maybePop(),
              icon: const Icon(Icons.close),
            ),
        ],
      ),
      content: SizedBox(
        width: maxWidth,
        child: ConstrainedBox(
          constraints: BoxConstraints(maxHeight: maxHeight),
          child: SingleChildScrollView(child: child),
        ),
      ),
    );
  }
}
