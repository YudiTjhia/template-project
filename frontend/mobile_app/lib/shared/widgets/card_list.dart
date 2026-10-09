import 'package:flutter/material.dart';

/// Renders a collection of items as separated cards.
///
/// By default the list is shrink-wrapped and does not scroll independently,
/// making it suitable for use inside a page's `ListView` or `CustomScrollView`.
class CardList<T> extends StatelessWidget {
  const CardList({
    super.key,
    required this.items,
    required this.itemBuilder,
    this.padding = EdgeInsets.zero,
    this.spacing = 12,
    this.cardMargin = EdgeInsets.zero,
    this.shrinkWrap = true,
    this.physics = const NeverScrollableScrollPhysics(),
    this.emptyBuilder,
  });

  final List<T> items;
  final Widget Function(BuildContext context, T item, int index) itemBuilder;
  final EdgeInsetsGeometry padding;
  final double spacing;
  final EdgeInsetsGeometry cardMargin;
  final bool shrinkWrap;
  final ScrollPhysics physics;
  final WidgetBuilder? emptyBuilder;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return emptyBuilder?.call(context) ?? const SizedBox.shrink();
    }

    return ListView.separated(
      padding: padding,
      shrinkWrap: shrinkWrap,
      physics: physics,
      itemCount: items.length,
      separatorBuilder: (_, _) => SizedBox(height: spacing),
      itemBuilder: (context, index) => Card(
        margin: cardMargin,
        clipBehavior: Clip.antiAlias,
        child: itemBuilder(context, items[index], index),
      ),
    );
  }
}
