import 'package:flutter/material.dart';

import '../config/app_config.dart';

/// Reusable register form with configurable application branding.
class RegisterComponent extends StatefulWidget {
  const RegisterComponent({
    super.key,
    required this.formKey,
    required this.confirmPasswordFieldKey,
    required this.nameController,
    required this.emailController,
    required this.phoneController,
    required this.passwordController,
    required this.confirmPasswordController,
    required this.isLoading,
    required this.onRegister,
    required this.onLogin,
    required this.validatePassword,
    required this.validateConfirmPassword,
    this.appTitle,
    this.logo,
  });

  final GlobalKey<FormState> formKey;
  final GlobalKey<FormFieldState<String>> confirmPasswordFieldKey;
  final TextEditingController nameController;
  final TextEditingController emailController;
  final TextEditingController phoneController;
  final TextEditingController passwordController;
  final TextEditingController confirmPasswordController;
  final bool isLoading;
  final VoidCallback onRegister;
  final VoidCallback onLogin;
  final FormFieldValidator<String> validatePassword;
  final FormFieldValidator<String> validateConfirmPassword;
  final String? appTitle;
  final ImageProvider<Object>? logo;

  @override
  State<RegisterComponent> createState() => _RegisterComponentState();
}

class _RegisterComponentState extends State<RegisterComponent> {
  static const _brandColor = Color(0xFF2196F3);
  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Form(
        key: widget.formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildLogo(),
              const SizedBox(height: 12),
              Text(
                widget.appTitle ?? AppConfig.appTitle,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: _brandColor,
                  fontSize: 24,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 28),
              TextFormField(
                controller: widget.nameController,
                enabled: !widget.isLoading,
                textInputAction: TextInputAction.next,
                decoration: const InputDecoration(
                  labelText: 'Full Name',
                  prefixIcon: Icon(Icons.person_outline),
                  border: OutlineInputBorder(),
                ),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Full name wajib diisi';
                  }
                  if (value.trim().length < 3) return 'Nama minimal 3 karakter';
                  return null;
                },
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: widget.emailController,
                enabled: !widget.isLoading,
                keyboardType: TextInputType.emailAddress,
                textInputAction: TextInputAction.next,
                decoration: const InputDecoration(
                  labelText: 'Email',
                  prefixIcon: Icon(Icons.email_outlined),
                  border: OutlineInputBorder(),
                ),
                validator: _validateEmail,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: widget.phoneController,
                enabled: !widget.isLoading,
                keyboardType: TextInputType.phone,
                textInputAction: TextInputAction.next,
                decoration: const InputDecoration(
                  labelText: 'Phone Number',
                  hintText: '08xxxxxxxxxx',
                  prefixIcon: Icon(Icons.phone_outlined),
                  border: OutlineInputBorder(),
                ),
                validator: _validatePhone,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: widget.passwordController,
                enabled: !widget.isLoading,
                obscureText: _obscurePassword,
                textInputAction: TextInputAction.next,
                decoration: InputDecoration(
                  labelText: 'Password',
                  prefixIcon: const Icon(Icons.lock_outline),
                  border: const OutlineInputBorder(),
                  suffixIcon: IconButton(
                    onPressed: widget.isLoading
                        ? null
                        : () => setState(
                            () => _obscurePassword = !_obscurePassword,
                          ),
                    icon: Icon(
                      _obscurePassword
                          ? Icons.visibility_outlined
                          : Icons.visibility_off_outlined,
                    ),
                  ),
                ),
                validator: widget.validatePassword,
              ),
              const SizedBox(height: 16),
              TextFormField(
                key: widget.confirmPasswordFieldKey,
                controller: widget.confirmPasswordController,
                enabled: !widget.isLoading,
                obscureText: _obscureConfirmPassword,
                textInputAction: TextInputAction.done,
                onFieldSubmitted: (_) => widget.onRegister(),
                decoration: InputDecoration(
                  labelText: 'Confirm Password',
                  prefixIcon: const Icon(Icons.lock_outline),
                  border: const OutlineInputBorder(),
                  suffixIcon: IconButton(
                    onPressed: widget.isLoading
                        ? null
                        : () => setState(
                            () => _obscureConfirmPassword =
                                !_obscureConfirmPassword,
                          ),
                    icon: Icon(
                      _obscureConfirmPassword
                          ? Icons.visibility_outlined
                          : Icons.visibility_off_outlined,
                    ),
                  ),
                ),
                validator: widget.validateConfirmPassword,
              ),
              const SizedBox(height: 24),
              SizedBox(
                height: 52,
                child: FilledButton(
                  onPressed: widget.isLoading ? null : widget.onRegister,
                  child: widget.isLoading
                      ? const SizedBox(
                          width: 22,
                          height: 22,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Daftar', style: TextStyle(fontSize: 16)),
                ),
              ),
              const SizedBox(height: 12),
              TextButton(
                onPressed: widget.isLoading ? null : widget.onLogin,
                child: const Text('Sudah punya akun? Login'),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildLogo() {
    final image = widget.logo;
    if (image == null) {
      return const Icon(Icons.point_of_sale, size: 72, color: _brandColor);
    }
    return Image(
      image: image,
      width: 72,
      height: 72,
      fit: BoxFit.contain,
      errorBuilder: (_, _, _) =>
          const Icon(Icons.point_of_sale, size: 72, color: _brandColor),
    );
  }

  String? _validateEmail(String? value) {
    final email = value?.trim() ?? '';
    if (email.isEmpty) return 'Email wajib diisi';
    if (!RegExp(
      r'^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$',
      caseSensitive: false,
    ).hasMatch(email)) {
      return 'Format email tidak valid';
    }
    return null;
  }

  String? _validatePhone(String? value) {
    if (value == null || value.trim().isEmpty) return 'Nomor HP wajib diisi';
    final normalized = value.trim().replaceAll(RegExp(r'[\s()-]'), '');
    if (!RegExp(r'^\+?[0-9]{9,15}$').hasMatch(normalized)) {
      return 'Nomor HP tidak valid';
    }
    return null;
  }
}
