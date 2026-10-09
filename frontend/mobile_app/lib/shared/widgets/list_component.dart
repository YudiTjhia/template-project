import 'package:flutter/material.dart';

class ListFieldMetadata<T> {
  const ListFieldMetadata({
    required this.label,
    required this.valueBuilder,
    this.icon,
    this.labelStyle,
    this.valueStyle,
    this.maxLines = 2,
  });

  final String label;
  final String Function(T item) valueBuilder;
  final IconData? icon;
  final TextStyle? labelStyle;
  final TextStyle? valueStyle;
  final int maxLines;
}

/// Builds a list using field metadata or a custom item builder.
class ListComponent<T> extends StatelessWidget {
  const ListComponent({
    super.key,
    required this.items,
    this.metadata = const [],
    this.itemBuilder,
    this.separatorBuilder,
    this.emptyBuilder,
    this.padding = EdgeInsets.zero,
    this.shrinkWrap = true,
    this.physics = const NeverScrollableScrollPhysics(),
    this.scrollDirection = Axis.vertical,
    this.cardItems = true,
  });

  final List<T> items;
  final List<ListFieldMetadata<T>> metadata;
  final Widget Function(BuildContext context, T item, int index)? itemBuilder;
  final IndexedWidgetBuilder? separatorBuilder;
  final WidgetBuilder? emptyBuilder;
  final EdgeInsetsGeometry padding;
  final bool shrinkWrap;
  final ScrollPhysics physics;
  final Axis scrollDirection;
  final bool cardItems;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) {
      return emptyBuilder?.call(context) ?? const SizedBox.shrink();
    }

    if (itemBuilder == null && metadata.isEmpty) {
      throw FlutterError(
        'ListComponent requires itemBuilder or at least one metadata field.',
      );
    }

    Widget buildRow(BuildContext context, int index) {
      final content =
          itemBuilder?.call(context, items[index], index) ??
          _buildMetadataItem(items[index]);
      if (!cardItems) return content;
      return Card(margin: EdgeInsets.zero, child: content);
    }

    final separator = separatorBuilder;
    if (separator == null) {
      return ListView.builder(
        padding: padding,
        shrinkWrap: shrinkWrap,
        physics: physics,
        scrollDirection: scrollDirection,
        itemCount: items.length,
        itemBuilder: buildRow,
      );
    }

    return ListView.separated(
      padding: padding,
      shrinkWrap: shrinkWrap,
      physics: physics,
      scrollDirection: scrollDirection,
      itemCount: items.length,
      separatorBuilder: separator,
      itemBuilder: buildRow,
    );
  }

  Widget _buildMetadataItem(T item) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          for (var index = 0; index < metadata.length; index++) ...[
            if (index > 0) const SizedBox(height: 12),
            _buildMetadataField(metadata[index], item),
          ],
        ],
      ),
    );
  }

  Widget _buildMetadataField(ListFieldMetadata<T> field, T item) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        if (field.icon != null) ...[
          Icon(field.icon, size: 20),
          const SizedBox(width: 10),
        ],
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(field.label, style: field.labelStyle),
              const SizedBox(height: 2),
              Text(
                field.valueBuilder(item),
                style: field.valueStyle,
                maxLines: field.maxLines,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
      ],
    );
  }
}
