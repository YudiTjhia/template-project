import 'package:flutter/material.dart';

import '../../shared/theme/app_theme.dart';
import '../../shared/widgets/msg_box.dart';
import '../../navigation/app_drawer_menu.dart';

import 'package:mobile_app/models/profile_model.dart';
import 'package:mobile_app/services/profile_service.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({
    super.key,
    this.enableDrawer = true,
    this.selectedRoute = '/profile/user',
  });

  final bool enableDrawer;
  final String selectedRoute;

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final _profileService = ProfileService();
  ProfileModel? _profile;
  String? _errorMessage;
  bool _isLoading = true;
  bool _isSavingProfile = false;

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final profile = await _profileService.getCurrentUser();
      if (!mounted) return;
      setState(() => _profile = profile);
    } on ProfileServiceException catch (error) {
      if (!mounted) return;
      setState(() => _errorMessage = error.message);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _editProfile() async {
    final profile = _profile;
    if (profile == null || _isSavingProfile) return;

    final formKey = GlobalKey<FormState>();
    final usernameController = TextEditingController(text: profile.username);
    final emailController = TextEditingController(text: profile.email);
    final shouldSave = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Edit profil'),
        content: Form(
          key: formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: usernameController,
                autofocus: true,
                textInputAction: TextInputAction.next,
                decoration: const InputDecoration(labelText: 'Username'),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Username wajib diisi';
                  }
                  if (value.trim().length < 3) {
                    return 'Username minimal 3 karakter';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: emailController,
                keyboardType: TextInputType.emailAddress,
                textInputAction: TextInputAction.done,
                decoration: const InputDecoration(labelText: 'Email'),
                validator: (value) {
                  final email = value?.trim() ?? '';
                  if (email.isEmpty) return 'Email wajib diisi';
                  if (!RegExp(
                    r'^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$',
                    caseSensitive: false,
                  ).hasMatch(email)) {
                    return 'Format email tidak valid';
                  }
                  return null;
                },
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext, false),
            child: const Text('Batal'),
          ),
          FilledButton(
            onPressed: () {
              if (formKey.currentState?.validate() ?? false) {
                Navigator.pop(dialogContext, true);
              }
            },
            child: const Text('Simpan'),
          ),
        ],
      ),
    );

    if (shouldSave != true || !mounted) {
      usernameController.dispose();
      emailController.dispose();
      return;
    }

    setState(() => _isSavingProfile = true);
    try {
      await _profileService.updateCurrentUser(
        username: usernameController.text,
        email: emailController.text,
      );
      await _loadProfile();
      if (!mounted) return;
      await MsgBox.show(
        context,
        title: 'Profil diperbarui',
        message: 'Perubahan profil berhasil disimpan.',
      );
    } on ProfileServiceException catch (error) {
      if (!mounted) return;
      await MsgBox.show(
        context,
        title: 'Gagal memperbarui profil',
        message: error.message,
        buttonText: 'Tutup',
      );
    } finally {
      usernameController.dispose();
      emailController.dispose();
      if (mounted) setState(() => _isSavingProfile = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.pageBackgroundColor,
      appBar: AppBar(title: const Text('Profile')),
      drawer: widget.enableDrawer
          ? AppDrawerMenu(selectedRoute: widget.selectedRoute)
          : null,
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [_buildProfileCard()],
        ),
      ),
    );
  }

  Widget _buildProfileCard() {
    if (_isLoading) {
      return const Card(
        child: SizedBox(
          height: 220,
          child: Center(child: CircularProgressIndicator()),
        ),
      );
    }

    if (_profile == null) {
      return Card(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            children: [
              const Icon(Icons.person_off_outlined, size: 42),
              const SizedBox(height: 8),
              Text(_errorMessage ?? 'Profil tidak tersedia'),
              const SizedBox(height: 8),
              TextButton.icon(
                onPressed: _loadProfile,
                icon: const Icon(Icons.refresh),
                label: const Text('Coba lagi'),
              ),
            ],
          ),
        ),
      );
    }

    final profile = _profile!;
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            _ProfileAvatar(profile: profile),
            const SizedBox(height: 14),
            Text(
              profile.name.isEmpty ? 'User' : profile.name,
              textAlign: TextAlign.center,
              style: const TextStyle(fontSize: 20, fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 20),
            _ProfileField(label: 'Username', value: profile.username),
            const Divider(height: 24),
            _ProfileField(label: 'Email', value: profile.email),
            const SizedBox(height: 16),
            OutlinedButton.icon(
              onPressed: _isSavingProfile ? null : _editProfile,
              icon: _isSavingProfile
                  ? const SizedBox.square(
                      dimension: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.edit_outlined),
              label: const Text('Edit profil'),
            ),
          ],
        ),
      ),
    );
  }
}

class _ProfileAvatar extends StatelessWidget {
  const _ProfileAvatar({required this.profile});

  final ProfileModel profile;

  @override
  Widget build(BuildContext context) {
    final initials = profile.name.trim().isEmpty
        ? '?'
        : profile.name.trim().characters.first.toUpperCase();
    final photoUrl = profile.photoUrl?.trim() ?? '';

    return CircleAvatar(
      radius: 48,
      backgroundColor: Theme.of(context).colorScheme.primaryContainer,
      child: ClipOval(
        child: photoUrl.isEmpty
            ? _initials(initials)
            : Image.network(
                photoUrl,
                width: 96,
                height: 96,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) => _initials(initials),
              ),
      ),
    );
  }

  Widget _initials(String value) => SizedBox(
    width: 96,
    height: 96,
    child: Center(
      child: Text(
        value,
        style: const TextStyle(fontSize: 32, fontWeight: FontWeight.w700),
      ),
    ),
  );
}

class _ProfileField extends StatelessWidget {
  const _ProfileField({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          width: 92,
          child: Text(label, style: const TextStyle(color: Color(0xFF667085))),
        ),
        Expanded(
          child: Text(
            value.isEmpty ? '—' : value,
            style: const TextStyle(
              color: Color(0xFF101828),
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
      ],
    );
  }
}
