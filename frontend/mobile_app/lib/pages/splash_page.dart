import 'dart:async';

import 'package:flutter/material.dart';

import 'home_page.dart';
import 'login_page.dart';
import '../services/login_service.dart';
import '../shared/helpers/session_helper.dart';

const _splashBackground = Color(0xFFB9E3FF);
const _splashForeground = Color(0xFF12324A);

class SplashPage extends StatefulWidget {
  const SplashPage({super.key, this.logo, this.appName = 'Open POS'});

  final ImageProvider<Object>? logo;
  final String appName;

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
        _openPage(const LoginPage());
        return;
      }

      final isSessionValid = await LoginService().validateSession();
      if (!mounted) return;
      _openPage(isSessionValid ? const HomePage() : const LoginPage());
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
      body: Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (widget.logo == null)
              const Icon(
                Icons.point_of_sale,
                size: 88,
                color: _splashForeground,
              )
            else
              Image(
                image: widget.logo!,
                width: 88,
                height: 88,
                fit: BoxFit.contain,
                errorBuilder: (_, _, _) => const Icon(
                  Icons.point_of_sale,
                  size: 88,
                  color: _splashForeground,
                ),
              ),
            const SizedBox(height: 16),
            Text(
              widget.appName,
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                fontWeight: FontWeight.bold,
                color: _splashForeground,
              ),
            ),
            const SizedBox(height: 28),
            if (isChecking)
              const CircularProgressIndicator(
                color: _splashForeground,
                backgroundColor: Color(0x6684B9D9),
              )
            else
              Column(
                children: [
                  Text(
                    errorMessage ?? 'Gagal memvalidasi sesi',
                    textAlign: TextAlign.center,
                    style: const TextStyle(color: _splashForeground),
                  ),
                  const SizedBox(height: 12),
                  TextButton(
                    onPressed: _checkSession,
                    style: TextButton.styleFrom(
                      foregroundColor: _splashForeground,
                    ),
                    child: const Text('Coba lagi'),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }
}
