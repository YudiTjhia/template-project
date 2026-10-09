import 'package:flutter/material.dart';

import '../shared/navigation/page_navigation.dart';
import '../shared/widgets/bottom_menu.dart';

class AppBottomMenu extends StatelessWidget {
  const AppBottomMenu({
    super.key,
    required this.selected,
    this.items = defaultItems,
    this.onMenuSelected,
  });

  final MenuDestination selected;
  final List<MenuItem> items;
  final ValueChanged<MenuItem>? onMenuSelected;

  static const defaultItems = [
    MenuItem(
      destination: MenuDestination.home,
      label: 'Home',
      icon: Icons.home_outlined,
      selectedIcon: Icons.home,
      route: '/home',
    ),
    MenuItem(
      destination: MenuDestination.search,
      label: 'Search',
      icon: Icons.search_outlined,
      selectedIcon: Icons.search,
      route: '/search',
    ),
    MenuItem(
      destination: MenuDestination.aiChat,
      label: 'AI Chat',
      icon: Icons.auto_awesome_outlined,
      selectedIcon: Icons.auto_awesome,
      route: '/ai-chat',
      isCenter: true,
    ),
    MenuItem(
      destination: MenuDestination.carts,
      label: 'Carts',
      icon: Icons.shopping_cart_outlined,
      selectedIcon: Icons.shopping_cart,
      route: '/carts',
    ),
    MenuItem(
      destination: MenuDestination.history,
      label: 'History',
      icon: Icons.receipt_long_outlined,
      selectedIcon: Icons.receipt_long,
      route: '/history',
    ),
    MenuItem(
      destination: MenuDestination.profile,
      label: 'Profile',
      icon: Icons.person_outline,
      selectedIcon: Icons.person,
      route: '/profile',
    ),
  ];

  @override
  Widget build(BuildContext context) {
    return BottomMenu(
      selected: selected,
      items: items,
      onSelected: (destination) {
        final matchingItems = items.where(
          (item) => item.destination == destination,
        );
        if (matchingItems.isEmpty) return;

        final item = matchingItems.first;
        if (onMenuSelected != null) {
          onMenuSelected!(item);
        } else {
          PageNavigation.go(context, item, current: selected);
        }
      },
    );
  }
}
