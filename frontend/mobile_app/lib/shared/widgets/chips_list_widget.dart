import 'package:flutter/material.dart';

/// Displays values as a wrapping collection of chips.
///
/// Use [labelBuilder] when items are models. Without it, each item is converted
/// to a string. Optional callbacks add selection and remove actions.
class ChipsListWidget<T> extends StatelessWidget {
  const ChipsListWidget({
    super.key,
    required this.items,
    this.labelBuilder,
    this.avatarBuilder,
    this.onSelected,
    this.onDeleted,
    this.isSelected,
    this.tooltipBuilder,
    this.emptyWidget,
    this.spacing = 8,
    this.runSpacing = 8,
    this.alignment = WrapAlignment.start,
    this.chipColor,
    this.outlined = false,
  });

  final List<T> items;
  final String Function(T item)? labelBuilder;
  final Widget? Function(T item)? avatarBuilder;
  final ValueChanged<T>? onSelected;
  final ValueChanged<T>? onDeleted;
  final bool Function(T item)? isSelected;
  final String? Function(T item)? tooltipBuilder;
  final Widget? emptyWidget;
  final double spacing;
  final double runSpacing;
  final WrapAlignment alignment;
  final Color? chipColor;
  final bool outlined;

  @override
  Widget build(BuildContext context) {
    if (items.isEmpty) return emptyWidget ?? const SizedBox.shrink();

    return Wrap(
      spacing: spacing,
      runSpacing: runSpacing,
      alignment: alignment,
      children: [
        for (final item in items)
          InputChip(
            label: Text(labelBuilder?.call(item) ?? item.toString()),
            avatar: avatarBuilder?.call(item),
            selected: isSelected?.call(item) ?? false,
            showCheckmark: isSelected != null,
            tooltip: tooltipBuilder?.call(item),
            onPressed: onSelected == null ? null : () => onSelected!(item),
            onDeleted: onDeleted == null ? null : () => onDeleted!(item),
            backgroundColor: chipColor,
            side: outlined ? const BorderSide() : null,
          ),
      ],
    );
  }
}
