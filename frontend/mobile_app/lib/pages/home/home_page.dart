import 'package:flutter/material.dart';

import '../../shared/navigation/page_navigation.dart';
import '../../config/app_config.dart';
import '../../shared/theme/app_theme.dart';
import '../../navigation/app_drawer_menu.dart';
import '../../navigation/app_bottom_menu.dart';
import '../../shared/widgets/bottom_menu.dart';

class HomePage extends StatelessWidget {
  const HomePage({
    super.key,
    this.appTitle,
    this.useBottomMenu = true,
    this.enableDrawer = true,
    this.items = AppBottomMenu.defaultItems,
    this.onMenuSelected,
  });

  final String? appTitle;
  final bool useBottomMenu;
  final bool enableDrawer;
  final List<MenuItem> items;
  final ValueChanged<MenuItem>? onMenuSelected;

  @override
  Widget build(BuildContext context) {
    final title = appTitle ?? AppConfig.appTitle;
    return Scaffold(
      backgroundColor: AppTheme.pageBackgroundColor,
      appBar: AppBar(title: Text(title)),
      drawer: enableDrawer
          ? AppDrawerMenu(selectedRoute: '/home', headerTitle: title)
          : null,
      body: SafeArea(
        child: CustomScrollView(
          slivers: [
            SliverPadding(
              padding: const EdgeInsets.fromLTRB(20, 24, 20, 20),
              sliver: SliverToBoxAdapter(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      title,
                      style: Theme.of(context).textTheme.headlineSmall
                          ?.copyWith(
                            color: const Color(0xFF101828),
                            fontWeight: FontWeight.w800,
                          ),
                    ),
                    const SizedBox(height: 6),
                    const Text(
                      'Selamat datang. Apa yang ingin Anda lakukan hari ini?',
                      style: TextStyle(color: Color(0xFF667085)),
                    ),
                    const SizedBox(height: 24),
                    _WelcomeCard(onBrowse: () => _selectDestination(context)),
                    const SizedBox(height: 28),
                    Text(
                      'Ringkasan',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w800,
                        color: const Color(0xFF101828),
                      ),
                    ),
                    const SizedBox(height: 12),
                    const _EmptySummaryCard(),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: useBottomMenu
          ? AppBottomMenu(
              selected: MenuDestination.home,
              items: items,
              onMenuSelected: onMenuSelected,
            )
          : null,
    );
  }

  void _selectDestination(BuildContext context) {
    final destination = items.where(
      (item) => item.destination == MenuDestination.search,
    );
    if (destination.isEmpty) return;
    if (onMenuSelected != null) {
      onMenuSelected!(destination.first);
    } else {
      PageNavigation.go(
        context,
        destination.first,
        current: MenuDestination.home,
      );
    }
  }
}

class _WelcomeCard extends StatelessWidget {
  const _WelcomeCard({required this.onBrowse});

  final VoidCallback onBrowse;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.primary,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Icon(Icons.point_of_sale, color: Colors.white, size: 32),
          const SizedBox(height: 16),
          const Text(
            'Mulai kelola bisnis Anda',
            style: TextStyle(
              color: Colors.white,
              fontSize: 20,
              fontWeight: FontWeight.w800,
            ),
          ),
          const SizedBox(height: 6),
          const Text(
            'Cari produk dan lanjutkan aktivitas dari menu di bawah.',
            style: TextStyle(color: Color(0xFFEAF4FF)),
          ),
          const SizedBox(height: 18),
          FilledButton.tonalIcon(
            onPressed: onBrowse,
            icon: const Icon(Icons.search),
            label: const Text('Cari produk'),
          ),
        ],
      ),
    );
  }
}

class _EmptySummaryCard extends StatelessWidget {
  const _EmptySummaryCard();

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
      ),
      child: const Column(
        children: [
          Icon(Icons.dashboard_outlined, size: 36, color: Color(0xFF98A2B3)),
          SizedBox(height: 10),
          Text(
            'Belum ada ringkasan',
            style: TextStyle(
              color: Color(0xFF344054),
              fontWeight: FontWeight.w700,
            ),
          ),
          SizedBox(height: 4),
          Text(
            'Informasi aktivitas akan tampil di sini.',
            textAlign: TextAlign.center,
            style: TextStyle(color: Color(0xFF667085), fontSize: 13),
          ),
        ],
      ),
    );
  }
}
