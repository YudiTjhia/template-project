import 'package:flutter/material.dart';

class AppPageMenuItem {
  const AppPageMenuItem({
    required this.title,
    required this.icon,
    this.route,
    this.subtitle,
    this.children = const [],
    this.action,
  });

  final String title;
  final IconData icon;
  final String? route;
  final String? subtitle;
  final List<AppPageMenuItem> children;
  final AppPageMenuAction? action;

  bool get hasChildren => children.isNotEmpty;
}

/// Application pages that are available from the standard tile menu.
class AppPageNavigation {
  AppPageNavigation._();

  static const menuItems = <AppPageMenuItem>[
    AppPageMenuItem(
      title: 'Beranda',
      icon: Icons.home_outlined,
      route: '/home',
      subtitle: 'Kembali ke halaman utama',
    ),
    AppPageMenuItem(
      title: 'Komunikasi',
      icon: Icons.forum_outlined,
      subtitle: 'Notifikasi dan percakapan',
      children: [
        AppPageMenuItem(
          title: 'Notifikasi',
          icon: Icons.notifications_outlined,
          route: '/notifications',
          subtitle: 'Lihat pemberitahuan terbaru',
        ),
        AppPageMenuItem(
          title: 'AI Chat',
          icon: Icons.auto_awesome_outlined,
          route: '/ai-chat',
          subtitle: 'Buka percakapan AI',
        ),
      ],
    ),
    AppPageMenuItem(
      title: 'Profile',
      icon: Icons.account_circle_outlined,
      subtitle: 'Pengaturan account dan user profile',
      children: [
        AppPageMenuItem(
          title: 'Account Profile',
          icon: Icons.manage_accounts_outlined,
          route: '/profile/account',
        ),
        AppPageMenuItem(
          title: 'User Profile',
          icon: Icons.person_outline,
          route: '/profile/user',
        ),
        AppPageMenuItem(
          title: 'Change Password',
          icon: Icons.lock_outline,
          route: '/profile/change-password',
        ),
      ],
    ),
    AppPageMenuItem(
      title: 'User Management',
      icon: Icons.manage_accounts_outlined,
      subtitle: 'Kelola users dan roles',
      children: [
        AppPageMenuItem(
          title: 'Users',
          icon: Icons.people_outline,
          route: '/users',
          subtitle: 'Kelola akun pengguna',
        ),
        AppPageMenuItem(
          title: 'Roles',
          icon: Icons.admin_panel_settings_outlined,
          route: '/roles',
          subtitle: 'Atur role dan hak akses',
        ),
      ],
    ),
    AppPageMenuItem(
      title: 'Logout',
      icon: Icons.logout,
      action: AppPageMenuAction.logout,
    ),
  ];

  /// Returns all navigable items, including items nested inside groups.
  static List<AppPageMenuItem> get navigableItems {
    final result = <AppPageMenuItem>[];
    void collect(List<AppPageMenuItem> source) {
      for (final item in source) {
        if (item.route != null && item.route!.isNotEmpty) result.add(item);
        collect(item.children);
      }
    }

    collect(menuItems);
    return result;
  }

  static Future<T?> go<T extends Object?>(
    BuildContext context,
    String route, {
    bool replace = false,
  }) {
    if (replace) {
      return Navigator.of(context).pushReplacementNamed<T, Object?>(route);
    }
    return Navigator.of(context).pushNamed<T>(route);
  }
}

enum AppPageMenuAction { logout }
