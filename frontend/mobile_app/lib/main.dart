import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';

import 'pages/home/home_page.dart';
import 'pages/login/splash_page.dart';
import 'pages/notifications/notification_page.dart';
import 'pages/roles/role_edit_page.dart';
import 'pages/roles/roles_list_page.dart';
import 'pages/users/profile_page.dart';
import 'pages/users/user_edit_page.dart';
import 'pages/users/users_list_page.dart';
import 'services/notification_service.dart';
import 'shared/helpers/firebase_options.dart';
import 'shared/theme/app_theme.dart';

@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
}

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  if (!kIsWeb && defaultTargetPlatform == TargetPlatform.android) {
    await Firebase.initializeApp(
      options: DefaultFirebaseOptions.currentPlatform,
    );
    FirebaseMessaging.onBackgroundMessage(firebaseMessagingBackgroundHandler);
    await NotificationService.instance.initialize();
  }
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
      routes: {
        '/home': (_) => const HomePage(),
        '/profile': (_) => const ProfilePage(),
        '/notifications': (_) => const NotificationPage(),
        '/users': (_) => const UsersListPage(),
        '/users/edit': (context) {
          final userId = ModalRoute.of(context)?.settings.arguments as int?;
          return UserEditPage(userId: userId);
        },
        '/roles': (_) => const RolesListPage(),
        '/roles/edit': (context) {
          final roleId = ModalRoute.of(context)?.settings.arguments as int?;
          return RoleEditPage(roleId: roleId);
        },
      },
      home: const SplashPage(),
    );
  }
}
