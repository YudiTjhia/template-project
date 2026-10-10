import 'package:flutter/material.dart';

import '../navigation/app_drawer_menu.dart';
import '../services/login_service.dart';
import '../shared/widgets/msg_box.dart';
import '../shared/widgets/responsive_page_layout.dart';

class ChangePasswordPage extends StatefulWidget {
  const ChangePasswordPage({super.key, this.service, this.enableDrawer = true});

  final LoginService? service;
  final bool enableDrawer;

  @override
  State<ChangePasswordPage> createState() => _ChangePasswordPageState();
}

class _ChangePasswordPageState extends State<ChangePasswordPage> {
  late final LoginService _service = widget.service ?? LoginService();
  final _formKey = GlobalKey<FormState>();
  final _currentPassword = TextEditingController();
  final _newPassword = TextEditingController();
  final _confirmPassword = TextEditingController();
  bool _saving = false;
  bool _obscureCurrent = true;
  bool _obscureNew = true;
  bool _obscureConfirm = true;

  @override
  void dispose() {
    _currentPassword.dispose();
    _newPassword.dispose();
    _confirmPassword.dispose();
    super.dispose();
  }

  Future<void> _changePassword() async {
    if (_saving || !(_formKey.currentState?.validate() ?? false)) return;
    setState(() => _saving = true);
    try {
      await _service.changePassword(
        currentPassword: _currentPassword.text,
        newPassword: _newPassword.text,
        confirmPassword: _confirmPassword.text,
      );
      if (!mounted) return;
      _currentPassword.clear();
      _newPassword.clear();
      _confirmPassword.clear();
      await MsgBox.show(
        context,
        title: 'Password diubah',
        message: 'Password berhasil diperbarui.',
      );
    } on LoginServiceException catch (error) {
      if (!mounted) return;
      await MsgBox.show(
        context,
        title: 'Gagal mengubah password',
        message: error.message,
        buttonText: 'Tutup',
      );
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Change Password')),
      drawer: widget.enableDrawer
          ? const AppDrawerMenu(selectedRoute: '/profile/change-password')
          : null,
      body: ResponsivePageLayout(
        maxWidth: 720,
        safeArea: true,
        child: SafeArea(
          child: Form(
            key: _formKey,
            child: ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Text(
                  'Perbarui password account Anda.',
                  style: Theme.of(context).textTheme.bodyLarge,
                ),
                const SizedBox(height: 20),
                _passwordField(
                  controller: _currentPassword,
                  label: 'Password saat ini',
                  obscure: _obscureCurrent,
                  toggleObscure: () =>
                      setState(() => _obscureCurrent = !_obscureCurrent),
                  validator: (value) => value == null || value.isEmpty
                      ? 'Password saat ini wajib diisi'
                      : null,
                ),
                const SizedBox(height: 16),
                _passwordField(
                  controller: _newPassword,
                  label: 'Password baru',
                  obscure: _obscureNew,
                  toggleObscure: () =>
                      setState(() => _obscureNew = !_obscureNew),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Password baru wajib diisi';
                    }
                    if (value.length < 6) return 'Password minimal 6 karakter';
                    if (value == _currentPassword.text) {
                      return 'Password baru harus berbeda';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 16),
                _passwordField(
                  controller: _confirmPassword,
                  label: 'Konfirmasi password baru',
                  obscure: _obscureConfirm,
                  toggleObscure: () =>
                      setState(() => _obscureConfirm = !_obscureConfirm),
                  onSubmitted: (_) => _changePassword(),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Konfirmasi password wajib diisi';
                    }
                    if (value != _newPassword.text) {
                      return 'Konfirmasi password tidak sama';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 24),
                SizedBox(
                  height: 52,
                  child: FilledButton.icon(
                    onPressed: _saving ? null : _changePassword,
                    icon: _saving
                        ? const SizedBox.square(
                            dimension: 18,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.lock_reset_outlined),
                    label: const Text('Simpan password'),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _passwordField({
    required TextEditingController controller,
    required String label,
    required bool obscure,
    required VoidCallback toggleObscure,
    required FormFieldValidator<String> validator,
    ValueChanged<String>? onSubmitted,
  }) {
    return TextFormField(
      controller: controller,
      enabled: !_saving,
      obscureText: obscure,
      textInputAction: onSubmitted == null
          ? TextInputAction.next
          : TextInputAction.done,
      onFieldSubmitted: onSubmitted,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: const Icon(Icons.lock_outline),
        suffixIcon: IconButton(
          onPressed: _saving ? null : toggleObscure,
          icon: Icon(
            obscure ? Icons.visibility_outlined : Icons.visibility_off_outlined,
          ),
        ),
        border: const OutlineInputBorder(),
      ),
      validator: validator,
    );
  }
}
