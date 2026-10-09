import 'dart:async';

import 'package:flutter/material.dart';

import '../home/home_page.dart';
import 'login_page.dart';
import '../../services/login_service.dart';
import '../../config/app_config.dart';
import '../../shared/helpers/session_helper.dart';
import '../../widgets/splash_component.dart';

const _splashBackground = Color(0xFFB9E3FF);

class SplashPage extends StatefulWidget {
  const SplashPage({super.key, this.logo, this.appName});

  final ImageProvider<Object>? logo;
  final String? appName;

  String get _appTitle => appName ?? AppConfig.appTitle;

  @override
  State<SplashPage> createState() => _SplashPageState();
}

class _SplashPageState extends State<SplashPage> {
  @override
  void initState() {
    super.initState();
    unawaited(_checkSession(showDelay: true));
  }

  bool isChecking = true;
  String? errorMessage;

  Future<void> _checkSession({bool showDelay = false}) async {
    setState(() {
      isChecking = true;
      errorMessage = null;
    });
    if (showDelay) await Future<void>.delayed(const Duration(seconds: 2));

    try {
      final hasLocalSession = await SessionHelper.hasLoginSession();
      if (!hasLocalSession) {
        await LoginService().clearSession();
        _openPage(LoginPage(appTitle: widget._appTitle, logo: widget.logo));
        return;
      }

      final isSessionValid = await LoginService().validateSession();
      if (!mounted) return;
      _openPage(
        isSessionValid
            ? const HomePage()
            : LoginPage(appTitle: widget._appTitle, logo: widget.logo),
      );
    } on LoginServiceException catch (error) {
      if (!mounted) return;
      setState(() {
        isChecking = false;
        errorMessage = error.message;
      });
    }
  }

  void _openPage(Widget page) {
    if (!mounted) return;
    Navigator.of(context)
        .pushReplacement(MaterialPageRoute<void>(builder: (_) => page));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _splashBackground,
      body: SplashComponent(
        appTitle: widget._appTitle,
        logo: widget.logo,
        isLoading: isChecking,
        errorMessage: errorMessage ?? 'Gagal memvalidasi sesi',
        onRetry: _checkSession,
      ),
    );
  }
}
