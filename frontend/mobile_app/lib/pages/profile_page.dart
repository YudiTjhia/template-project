import 'package:flutter/material.dart';

import '../models/profile_model.dart';
import '../services/login_service.dart';
import '../services/profile_service.dart';
import '../shared/theme/app_theme.dart';
import 'login_page.dart';

class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final _profileService = ProfileService();
  ProfileModel? _profile;
  String? _errorMessage;
  bool _isLoading = true;
  bool _isLoggingOut = false;

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

  Future<void> _logout() async {
    final shouldLogout = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Logout'),
        content: const Text('Yakin ingin keluar dari akun ini?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Batal'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Logout'),
          ),
        ],
      ),
    );
    if (shouldLogout != true || !mounted) return;

    setState(() => _isLoggingOut = true);
    try {
      await LoginService().logout();
    } catch (_) {
      // The local session is cleared in LoginService even if the server is offline.
    }
    if (!mounted) return;
    Navigator.of(context).pushAndRemoveUntil(
      MaterialPageRoute<void>(builder: (_) => const LoginPage()),
      (_) => false,
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.pageBackgroundColor,
      appBar: AppBar(title: const Text('Profile')),
      body: SafeArea(
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            _buildProfileCard(),
            const SizedBox(height: 20),
            Card(
              child: ListTile(
                leading: const Icon(Icons.logout, color: Colors.red),
                title: const Text('Logout'),
                trailing: _isLoggingOut
                    ? const SizedBox.square(
                        dimension: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.chevron_right),
                enabled: !_isLoggingOut,
                onTap: _isLoggingOut ? null : _logout,
              ),
            ),
          ],
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
