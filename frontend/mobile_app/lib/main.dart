import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

import 'pages/home/home_page.dart';
import 'pages/feature_placeholder_page.dart';
import 'pages/account_profile_page.dart';
import 'pages/change_password_page.dart';
import 'pages/chat/chat_page.dart';
import 'pages/login/splash_page.dart';
import 'pages/notifications/notification_page.dart';
import 'pages/roles/roles_list_page.dart';
import 'pages/users/profile_page.dart';
import 'pages/users/user_edit_page.dart';
import 'pages/users/users_list_page.dart';
import 'services/notification_service.dart';
import 'services/dummy_user_service.dart';
import 'services/dummy_role_service.dart';
import 'config/app_config.dart';
import 'shared/helpers/firebase_options.dart';
import 'shared/theme/app_theme.dart';
import 'shared/widgets/bottom_menu.dart';

@pragma('vm:entry-point')
Future<void> firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp(options: DefaultFirebaseOptions.currentPlatform);
}

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await dotenv.load(fileName: '.env', isOptional: true);
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
  const MobileApp({super.key, this.requireLogin = false});

  /// Set to false for apps that should open directly to the home page.
  final bool requireLogin;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: AppConfig.appTitle,
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      routes: {
        '/home': (_) => const HomePage(),
        '/ai-chat': (_) => const ChatPage(),
        '/search': (_) => const FeaturePlaceholderPage(
          title: 'Search',
          icon: Icons.search_outlined,
          selectedDestination: MenuDestination.search,
        ),
        '/carts': (_) => const FeaturePlaceholderPage(
          title: 'Carts',
          icon: Icons.shopping_cart_outlined,
          selectedDestination: MenuDestination.carts,
        ),
        '/history': (_) => const FeaturePlaceholderPage(
          title: 'History',
          icon: Icons.receipt_long_outlined,
          selectedDestination: MenuDestination.history,
        ),
        '/profile': (_) => const ProfilePage(),
        '/profile/account': (_) => const AccountProfilePage(),
        '/profile/user': (_) => const ProfilePage(),
        '/profile/change-password': (_) => const ChangePasswordPage(),
        '/notifications': (_) => const NotificationPage(),
        '/users': (_) => UsersListPage(
          service: AppConfig.useDummyUsers ? DummyUserService() : null,
        ),
        '/users/edit': (context) {
          final userId = ModalRoute.of(context)?.settings.arguments as int?;
          return UserEditPage(userId: userId);
        },
        '/roles': (_) => RolesListPage(
          service: AppConfig.useDummyRoles ? DummyRoleService() : null,
        ),
      },
      home: requireLogin ? const SplashPage() : const HomePage(),
    );
  }
}
