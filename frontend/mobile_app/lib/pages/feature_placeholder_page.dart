import 'package:flutter/material.dart';

import '../navigation/app_bottom_menu.dart';
import '../navigation/app_drawer_menu.dart';
import '../shared/widgets/bottom_menu.dart';

/// Temporary destination used until a menu feature has its own page.
class FeaturePlaceholderPage extends StatelessWidget {
  const FeaturePlaceholderPage({
    super.key,
    required this.title,
    required this.icon,
    required this.selectedDestination,
  });

  final String title;
  final IconData icon;
  final MenuDestination selectedDestination;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(title)),
      drawer: const AppDrawerMenu(),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(
                icon,
                size: 56,
                color: Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 16),
              Text(
                title,
                style: Theme.of(context).textTheme.titleLarge,
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              const Text(
                'Halaman ini belum tersedia.',
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ),
      bottomNavigationBar: AppBottomMenu(selected: selectedDestination),
    );
  }
}
