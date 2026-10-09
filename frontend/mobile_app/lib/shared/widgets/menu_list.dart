import 'package:flutter/material.dart';

class MenuListItem {
  const MenuListItem({
    required this.title,
    this.icon,
    this.subtitle,
    this.trailing,
    this.onTap,
  });

  final String title;
  final IconData? icon;
  final String? subtitle;
  final Widget? trailing;
  final VoidCallback? onTap;
}

/// A grouped list of tappable menu rows.
class MenuList extends StatelessWidget {
  const MenuList({
    super.key,
    required this.items,
    this.backgroundColor,
    this.borderRadius = 16,
    this.dividerIndent = 56,
  });

  final List<MenuListItem> items;
  final Color? backgroundColor;
  final double borderRadius;
  final double dividerIndent;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: EdgeInsets.zero,
      color: backgroundColor ?? Theme.of(context).colorScheme.surface,
      clipBehavior: Clip.antiAlias,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(borderRadius),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          for (var index = 0; index < items.length; index++) ...[
            if (index > 0)
              Divider(height: 1, indent: dividerIndent, endIndent: 16),
            _MenuListRow(item: items[index]),
          ],
        ],
      ),
    );
  }
}

class _MenuListRow extends StatelessWidget {
  const _MenuListRow({required this.item});

  final MenuListItem item;

  @override
  Widget build(BuildContext context) {
    final icon = item.icon;
    return ListTile(
      leading: icon == null ? null : Icon(icon),
      title: Text(item.title),
      subtitle: item.subtitle == null ? null : Text(item.subtitle!),
      trailing: item.trailing ?? const Icon(Icons.chevron_right),
      onTap: item.onTap,
    );
  }
}
