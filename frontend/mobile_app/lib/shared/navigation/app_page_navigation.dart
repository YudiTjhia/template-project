import 'package:flutter/material.dart';

class AppPageMenuItem {
  const AppPageMenuItem({
    required this.title,
    required this.icon,
    required this.route,
    this.subtitle,
  });

  final String title;
  final IconData icon;
  final String route;
  final String? subtitle;
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
      title: 'Notifikasi',
      icon: Icons.notifications_outlined,
      route: '/notifications',
      subtitle: 'Lihat pemberitahuan terbaru',
    ),
    AppPageMenuItem(
      title: 'Profil',
      icon: Icons.person_outline,
      route: '/profile',
      subtitle: 'Lihat dan ubah profil akun',
    ),
    AppPageMenuItem(
      title: 'User',
      icon: Icons.people_outline,
      route: '/users',
      subtitle: 'Kelola user dan role akun',
    ),
    AppPageMenuItem(
      title: 'Role',
      icon: Icons.admin_panel_settings_outlined,
      route: '/roles',
      subtitle: 'Atur role dan hak akses',
    ),
  ];

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
