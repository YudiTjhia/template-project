import 'package:flutter/material.dart';

import '../widgets/bottom_menu.dart';

class PageNavigation {
  static void go(
    BuildContext context,
    MenuItem item, {
    required MenuDestination current,
  }) {
    if (item.destination == current) return;
    final route = item.route;
    if (route == null || route.isEmpty) return;

    Navigator.of(context).pushReplacementNamed(route);
  }
}
