import 'package:flutter/material.dart';

/// Centers page content and adapts its width and spacing to the screen size.
///
/// Use it as the `Scaffold.body` for both scrollable content and full-height
/// layouts. The child receives the full available height, which keeps widgets
/// such as `Expanded` and `ListView` working as expected.
class ResponsivePageLayout extends StatelessWidget {
  const ResponsivePageLayout({
    super.key,
    required this.child,
    this.maxWidth = 1200,
    this.compactPadding = 16,
    this.mediumPadding = 24,
    this.expandedPadding = 32,
    this.safeArea = false,
  });

  final Widget child;
  final double maxWidth;
  final double compactPadding;
  final double mediumPadding;
  final double expandedPadding;
  final bool safeArea;

  @override
  Widget build(BuildContext context) {
    Widget layout = LayoutBuilder(
      builder: (context, constraints) {
        final width = constraints.maxWidth;
        final padding = width < 600
            ? compactPadding
            : width < 1024
            ? mediumPadding
            : expandedPadding;
        final boundedHeight = constraints.hasBoundedHeight;

        return Align(
          alignment: Alignment.topCenter,
          child: ConstrainedBox(
            constraints: BoxConstraints(maxWidth: maxWidth),
            child: SizedBox(
              width: double.infinity,
              height: boundedHeight ? constraints.maxHeight : null,
              child: Padding(
                padding: EdgeInsets.symmetric(horizontal: padding),
                child: child,
              ),
            ),
          ),
        );
      },
    );
    if (safeArea) layout = SafeArea(child: layout);
    return layout;
  }
}
