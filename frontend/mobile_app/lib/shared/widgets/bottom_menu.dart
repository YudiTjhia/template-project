import 'package:flutter/material.dart';

enum MenuDestination { home, search, aiChat, carts, history, profile }

class MenuItem {
  const MenuItem({
    required this.destination,
    required this.label,
    required this.icon,
    required this.selectedIcon,
    this.route,
    this.isCenter = false,
  });

  final MenuDestination destination;
  final String label;
  final IconData icon;
  final IconData selectedIcon;
  final String? route;
  final bool isCenter;
}

class BottomMenu extends StatelessWidget {
  const BottomMenu({
    super.key,
    required this.selected,
    required this.onSelected,
    required this.items,
  });

  final MenuDestination selected;
  final ValueChanged<MenuDestination> onSelected;
  final List<MenuItem> items;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      top: false,
      child: Container(
        height: 78,
        padding: const EdgeInsets.fromLTRB(8, 8, 8, 10),
        decoration: BoxDecoration(
          color: Theme.of(context).colorScheme.surface,
          boxShadow: const [
            BoxShadow(
              color: Color(0x1A000000),
              blurRadius: 18,
              offset: Offset(0, -8),
            ),
          ],
        ),
        child: Row(
          children: [
            for (final item in items)
              Expanded(
                child: item.isCenter
                    ? _CenterMenuButton(
                        item: item,
                        isSelected: selected == item.destination,
                        onTap: () => onSelected(item.destination),
                      )
                    : _MenuButton(
                        item: item,
                        isSelected: selected == item.destination,
                        onTap: () => onSelected(item.destination),
                      ),
              ),
          ],
        ),
      ),
    );
  }
}

class _MenuButton extends StatelessWidget {
  const _MenuButton({
    required this.item,
    required this.isSelected,
    required this.onTap,
  });

  final MenuItem item;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final color = isSelected
        ? Theme.of(context).colorScheme.primary
        : const Color(0xFF667085);

    return InkWell(
      borderRadius: BorderRadius.circular(16),
      onTap: onTap,
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(isSelected ? item.selectedIcon : item.icon, color: color),
          const SizedBox(height: 4),
          Text(
            item.label,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: TextStyle(
              fontSize: 11,
              fontWeight: isSelected ? FontWeight.w800 : FontWeight.w600,
              color: color,
            ),
          ),
        ],
      ),
    );
  }
}

class _CenterMenuButton extends StatelessWidget {
  const _CenterMenuButton({
    required this.item,
    required this.isSelected,
    required this.onTap,
  });

  final MenuItem item;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return SizedBox.expand(
      child: Stack(
        clipBehavior: Clip.none,
        children: [
          Positioned(
            top: -20,
            left: 0,
            right: 0,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Material(
                  color: isSelected
                      ? colorScheme.primary
                      : const Color(0xFF101828),
                  elevation: 8,
                  shadowColor: Colors.black.withValues(alpha: 0.28),
                  shape: const CircleBorder(),
                  child: InkWell(
                    customBorder: const CircleBorder(),
                    onTap: onTap,
                    child: SizedBox.square(
                      dimension: 60,
                      child: Icon(
                        isSelected ? item.selectedIcon : item.icon,
                        color: Colors.white,
                        size: 27,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  item.label,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.w800,
                    color: isSelected
                        ? colorScheme.primary
                        : const Color(0xFF101828),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
