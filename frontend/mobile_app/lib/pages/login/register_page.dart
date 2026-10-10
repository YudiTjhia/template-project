import 'package:flutter/material.dart';

import '../../models/register_model.dart';
import '../../services/register_service.dart';
import '../../config/app_config.dart';
import '../../shared/widgets/confirm_box.dart';
import '../../shared/widgets/msg_box.dart';
import '../../widgets/register_component.dart';
import '../../shared/widgets/responsive_page_layout.dart';

const _registerBackground = Color(0xFFB9E3FF);

class RegisterPage extends StatefulWidget {
  const RegisterPage({super.key, this.appTitle, this.logo});

  final String? appTitle;
  final ImageProvider<Object>? logo;

  String get _appTitle => appTitle ?? AppConfig.appTitle;

  @override
  State<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends State<RegisterPage> {
  final formKey = GlobalKey<FormState>();
  final confirmPasswordFieldKey = GlobalKey<FormFieldState<String>>();
  final registerService = RegisterService();

  final nameController = TextEditingController();
  final emailController = TextEditingController();
  final phoneController = TextEditingController();
  final passwordController = TextEditingController();
  final confirmPasswordController = TextEditingController();

  bool isLoading = false;

  @override
  void initState() {
    super.initState();
    passwordController.addListener(validateConfirmPassword);
  }

  @override
  void dispose() {
    passwordController.removeListener(validateConfirmPassword);
    nameController.dispose();
    emailController.dispose();
    phoneController.dispose();
    passwordController.dispose();
    confirmPasswordController.dispose();
    super.dispose();
  }

  Future<void> register() async {
    if (isLoading) return;

    if (!formKey.currentState!.validate()) {
      return;
    }

    setState(() => isLoading = true);
    try {
      final response = await registerService.register(
        RegisterModel(
          name: nameController.text.trim(),
          email: emailController.text.trim(),
          phone: phoneController.text.trim(),
          password: passwordController.text,
        ),
      );
      if (!mounted) return;

      if (response.success) {
        final goToLogin = await ConfirmBox.show(
          context,
          title: 'Registrasi berhasil',
          message: 'Akun berhasil dibuat. Kembali ke halaman login?',
          confirmText: 'Ke Login',
        );
        if (goToLogin && mounted) Navigator.pop(context, response);
      } else {
        await MsgBox.show(
          context,
          title: 'Registrasi gagal',
          message: response.message,
          buttonText: 'Tutup',
        );
      }
    } on Exception catch (error) {
      if (!mounted) return;
      await MsgBox.show(
        context,
        title: 'Registrasi gagal',
        message: 'Terjadi kesalahan: $error',
        buttonText: 'Tutup',
      );
    } finally {
      if (mounted) setState(() => isLoading = false);
    }
  }

  void validateConfirmPassword() {
    if (confirmPasswordController.text.isEmpty) {
      return;
    }

    confirmPasswordFieldKey.currentState?.validate();
  }

  String? validatePassword(String? value) {
    if (value == null || value.isEmpty) {
      return 'Password wajib diisi';
    }

    if (value.length < 6) {
      return 'Password minimal 6 karakter';
    }

    return null;
  }

  String? validateConfirmPasswordField(String? value) {
    if (value == null || value.isEmpty) {
      return 'Konfirmasi password wajib diisi';
    }

    if (value != passwordController.text) {
      return 'Password dan konfirmasi password harus sama';
    }

    return null;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _registerBackground,
      appBar: AppBar(title: const Text('Registrasi Pelanggan')),
      body: ResponsivePageLayout(
        maxWidth: 560,
        compactPadding: 0,
        mediumPadding: 24,
        expandedPadding: 32,
        child: RegisterComponent(
          formKey: formKey,
          confirmPasswordFieldKey: confirmPasswordFieldKey,
          nameController: nameController,
          emailController: emailController,
          phoneController: phoneController,
          passwordController: passwordController,
          confirmPasswordController: confirmPasswordController,
          isLoading: isLoading,
          onRegister: register,
          onLogin: () => Navigator.pop(context),
          validatePassword: validatePassword,
          validateConfirmPassword: validateConfirmPasswordField,
          appTitle: widget._appTitle,
          logo: widget.logo,
        ),
      ),
    );
  }
}
