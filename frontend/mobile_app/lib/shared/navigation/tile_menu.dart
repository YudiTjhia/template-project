import 'package:flutter/material.dart';

class TileMenuItem {
  const TileMenuItem({
    required this.title,
    required this.icon,
    required this.onTap,
    this.subtitle,
    this.iconColor,
  });

  final String title;
  final IconData icon;
  final VoidCallback onTap;
  final String? subtitle;
  final Color? iconColor;
}

class TileMenu extends StatelessWidget {
  const TileMenu({
    super.key,
    required this.items,
    this.backgroundColor,
    this.borderRadius = 16,
  });

  final List<TileMenuItem> items;
  final Color? backgroundColor;
  final double borderRadius;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Card(
      margin: EdgeInsets.zero,
      color: backgroundColor ?? colorScheme.surface,
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(borderRadius),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          for (var index = 0; index < items.length; index++) ...[
            if (index > 0) const Divider(height: 1, indent: 56, endIndent: 16),
            _TileMenuRow(item: items[index]),
          ],
        ],
      ),
    );
  }
}

class _TileMenuRow extends StatelessWidget {
  const _TileMenuRow({required this.item});

  final TileMenuItem item;

  @override
  Widget build(BuildContext context) {
    final color = item.iconColor ?? Theme.of(context).colorScheme.primary;

    return ListTile(
      leading: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.10),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Icon(item.icon, color: color, size: 21),
      ),
      title: Text(item.title),
      subtitle: item.subtitle == null ? null : Text(item.subtitle!),
      trailing: const Icon(Icons.chevron_right),
      onTap: item.onTap,
    );
  }
}
