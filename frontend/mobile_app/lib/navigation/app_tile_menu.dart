import 'package:flutter/material.dart';

import 'app_page_navigation.dart';
import '../shared/navigation/tile_menu.dart';

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
        for (final item in _navigableItems)
          TileMenuItem(
            title: item.title,
            icon: item.icon,
            subtitle: item.subtitle,
            onTap: () {
              final route = item.route;
              if (route != null && route.isNotEmpty) {
                AppPageNavigation.go(context, route);
              }
            },
          ),
      ],
    );
  }

  List<AppPageMenuItem> get _navigableItems {
    final result = <AppPageMenuItem>[];
    void collect(List<AppPageMenuItem> source) {
      for (final item in source) {
        if (item.route != null && item.route!.isNotEmpty) result.add(item);
        collect(item.children);
      }
    }

    collect(items);
    return result;
  }
}
