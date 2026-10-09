import 'package:flutter/material.dart';

/// Branded splash presentation. Session checks and navigation stay in the page.
class SplashComponent extends StatelessWidget {
  const SplashComponent({
    super.key,
    required this.appTitle,
    required this.isLoading,
    this.logo,
    this.errorMessage,
    this.onRetry,
  });

  final String appTitle;
  final ImageProvider<Object>? logo;
  final bool isLoading;
  final String? errorMessage;
  final VoidCallback? onRetry;

  static const _foreground = Color(0xFF12324A);

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildLogo(),
          const SizedBox(height: 16),
          Text(
            appTitle,
            style: Theme.of(context).textTheme.headlineMedium
                ?.copyWith(fontWeight: FontWeight.bold, color: _foreground),
          ),
          const SizedBox(height: 28),
          if (isLoading)
            const CircularProgressIndicator(
              color: _foreground,
              backgroundColor: Color(0x6684B9D9),
            )
          else ...[
            Text(
              errorMessage ?? 'Gagal memuat aplikasi',
              textAlign: TextAlign.center,
              style: const TextStyle(color: _foreground),
            ),
            if (onRetry != null) ...[
              const SizedBox(height: 12),
              TextButton(
                onPressed: onRetry,
                style: TextButton.styleFrom(foregroundColor: _foreground),
                child: const Text('Coba lagi'),
              ),
            ],
          ],
        ],
      ),
    );
  }

  Widget _buildLogo() {
    final image = logo;
    if (image == null) {
      return const Icon(Icons.point_of_sale, size: 88, color: _foreground);
    }
    return Image(
      image: image,
      width: 88,
      height: 88,
      fit: BoxFit.contain,
      errorBuilder: (_, _, _) =>
          const Icon(Icons.point_of_sale, size: 88, color: _foreground),
    );
  }
}
