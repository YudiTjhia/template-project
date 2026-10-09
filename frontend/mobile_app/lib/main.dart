import 'package:flutter/material.dart';

import 'pages/profile_page.dart';
import 'pages/splash_page.dart';
import 'shared/theme/app_theme.dart';

void main() {
  runApp(const MobileApp());
}

class MobileApp extends StatelessWidget {
  const MobileApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'POS Mobile',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      routes: {'/profile': (_) => const ProfilePage()},
      home: const SplashPage(),
    );
  }
}
