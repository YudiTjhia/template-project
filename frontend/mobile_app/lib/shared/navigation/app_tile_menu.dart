import 'package:flutter/material.dart';

import 'app_page_navigation.dart';
import 'tile_menu.dart';

/// A tile menu populated with the application's standard page links.
class AppTileMenu extends StatelessWidget {
  const AppTileMenu({
    super.key,
    this.items = AppPageNavigation.menuItems,
    this.backgroundColor,
    this.borderRadius = 16,
  });

  final List<AppPageMenuItem> items;
  final Color? backgroundColor;
  final double borderRadius;

  @override
  Widget build(BuildContext context) {
    return TileMenu(
      backgroundColor: backgroundColor,
      borderRadius: borderRadius,
      items: [
        for (final item in items)
          TileMenuItem(
            title: item.title,
            icon: item.icon,
            subtitle: item.subtitle,
            onTap: () => AppPageNavigation.go(context, item.route),
          ),
      ],
    );
  }
}
