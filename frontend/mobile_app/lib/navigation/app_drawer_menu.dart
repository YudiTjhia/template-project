import 'package:flutter/material.dart';

import '../config/app_config.dart';
import '../pages/login/login_page.dart';
import '../services/login_service.dart';
import 'app_page_navigation.dart';

/// Shared navigation drawer populated with the application's standard pages.
class AppDrawerMenu extends StatelessWidget {
  const AppDrawerMenu({
    super.key,
    this.enabled = true,
    this.items = AppPageNavigation.menuItems,
    this.selectedRoute,
    this.headerTitle,
    this.headerSubtitle = 'Navigasi aplikasi',
  });

  final bool enabled;
  final List<AppPageMenuItem> items;
  final String? selectedRoute;
  final String? headerTitle;
  final String headerSubtitle;

  @override
  Widget build(BuildContext context) {
    if (!enabled) return const SizedBox.shrink();

    return Drawer(
      child: SafeArea(
        child: Column(
          children: [
            _DrawerHeader(
              title: headerTitle ?? AppConfig.appTitle,
              subtitle: headerSubtitle,
            ),
            Expanded(
              child: ListView(
                padding: const EdgeInsets.symmetric(vertical: 8),
                children: [
                  for (final item in items) _buildMenuItem(context, item),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildMenuItem(BuildContext context, AppPageMenuItem item) {
    if (item.hasChildren) {
      return ExpansionTile(
        key: PageStorageKey<String>('drawer-${item.title}'),
        leading: Icon(item.icon),
        title: Text(item.title),
        subtitle: item.subtitle == null ? null : Text(item.subtitle!),
        initiallyExpanded: _containsSelectedRoute(item),
        childrenPadding: const EdgeInsets.only(left: 16),
        children: [
          for (final child in item.children) _buildMenuItem(context, child),
        ],
      );
    }

    final route = item.route;
    final action = item.action;
    return ListTile(
      leading: Icon(item.icon),
      title: Text(item.title),
      subtitle: item.subtitle == null ? null : Text(item.subtitle!),
      selected: route != null && route == selectedRoute,
      iconColor: action == AppPageMenuAction.logout ? Colors.red : null,
      textColor: action == AppPageMenuAction.logout ? Colors.red : null,
      onTap: action == AppPageMenuAction.logout
          ? () => _logout(context)
          : route == null
          ? null
          : () {
              Navigator.of(context).pop();
              if (route != selectedRoute) AppPageNavigation.go(context, route);
            },
    );
  }

  Future<void> _logout(BuildContext context) async {
    final shouldLogout = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Logout'),
        content: const Text('Yakin ingin keluar dari akun ini?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Batal'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(dialogContext, true),
            child: const Text('Logout'),
          ),
        ],
      ),
    );
    if (shouldLogout != true || !context.mounted) return;

    try {
      await LoginService().logout();
    } catch (_) {
      // LoginService clears the local session even if the API is unavailable.
    }
    if (!context.mounted) return;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute<void>(
        builder: (_) => LoginPage(appTitle: AppConfig.appTitle),
      ),
      (_) => false,
    );
  }

  bool _containsSelectedRoute(AppPageMenuItem item) {
    if (item.route == selectedRoute) return true;
    return item.children.any(_containsSelectedRoute);
  }
}

class _DrawerHeader extends StatelessWidget {
  const _DrawerHeader({required this.title, required this.subtitle});

  final String title;
  final String subtitle;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.fromLTRB(20, 28, 20, 20),
      color: colorScheme.primary,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CircleAvatar(
            backgroundColor: colorScheme.onPrimary,
            foregroundColor: colorScheme.primary,
            child: const Icon(Icons.point_of_sale),
          ),
          const SizedBox(height: 14),
          Text(
            title,
            style: Theme.of(context).textTheme.titleLarge?.copyWith(
              color: colorScheme.onPrimary,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            subtitle,
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
              color: colorScheme.onPrimary.withValues(alpha: 0.82),
            ),
          ),
        ],
      ),
    );
  }
}
