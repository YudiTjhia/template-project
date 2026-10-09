import 'package:flutter/material.dart';
import 'package:mobile_app/pages/home/home_page.dart';

import '../../models/login_model.dart';
import '../../services/login_service.dart';
import '../../config/app_config.dart';
import '../../shared/widgets/msg_box.dart';
import '../../widgets/login_component.dart';
import 'register_page.dart';

const _loginBackground = Color(0xFFB9E3FF);

class LoginPage extends StatefulWidget {
  const LoginPage({super.key, this.appTitle, this.logo});

  final String? appTitle;
  final ImageProvider<Object>? logo;

  String get _appTitle => appTitle ?? AppConfig.appTitle;

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final formKey = GlobalKey<FormState>();
  final emailController = TextEditingController();
  final passwordController = TextEditingController();
  final loginService = LoginService();

  bool isLoading = false;

  Future<void> login() async {
    if (isLoading) return;

    if (!(formKey.currentState?.validate() ?? false)) return;

    setState(() => isLoading = true);
    try {
      await loginService.login(
        LoginModel(
          identifier: emailController.text.trim(),
          password: passwordController.text,
        ),
      );
      if (!mounted) return;
      Navigator.pushReplacement(
        context,
        MaterialPageRoute<void>(builder: (_) => const HomePage()),
      );
    } on LoginServiceException catch (error) {
      if (!mounted) return;
      await MsgBox.show(
        context,
        title: 'Login gagal',
        message: error.message,
        buttonText: 'Tutup',
      );
    } on Exception catch (error) {
      if (!mounted) return;
      await MsgBox.show(
        context,
        title: 'Login gagal',
        message: 'Terjadi kesalahan: $error',
        buttonText: 'Tutup',
      );
    } finally {
      if (mounted) setState(() => isLoading = false);
    }
  }

  @override
  void dispose() {
    emailController.dispose();
    passwordController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _loginBackground,
      body: LoginComponent(
        formKey: formKey,
        emailController: emailController,
        passwordController: passwordController,
        isLoading: isLoading,
        appTitle: widget._appTitle,
        logo: widget.logo,
        onLogin: login,
        onRegister: () {
          Navigator.push(
            context,
            MaterialPageRoute(
              builder: (_) =>
                  RegisterPage(appTitle: widget._appTitle, logo: widget.logo),
            ),
          );
        },
      ),
    );
  }
}
