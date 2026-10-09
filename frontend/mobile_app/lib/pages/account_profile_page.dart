import 'package:flutter/material.dart';

import '../models/account_profile_model.dart';
import '../navigation/app_drawer_menu.dart';
import '../services/account_profile_service.dart';

class AccountProfilePage extends StatefulWidget {
  const AccountProfilePage({super.key, this.service, this.enableDrawer = true});

  final AccountProfileService? service;
  final bool enableDrawer;

  @override
  State<AccountProfilePage> createState() => _AccountProfilePageState();
}

class _AccountProfilePageState extends State<AccountProfilePage> {
  late final AccountProfileService _service =
      widget.service ?? AccountProfileService();
  AccountProfileModel? _account;
  String? _error;
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadAccount();
  }

  Future<void> _loadAccount() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final account = await _service.getCurrentAccount();
      if (!mounted) return;
      setState(() {
        _account = account;
        _loading = false;
      });
    } on AccountProfileServiceException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _loading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final account = _account;
    return Scaffold(
      appBar: AppBar(title: const Text('Account Profile')),
      drawer: widget.enableDrawer
          ? const AppDrawerMenu(selectedRoute: '/profile/account')
          : null,
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
          ? Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(_error!, textAlign: TextAlign.center),
                    const SizedBox(height: 12),
                    OutlinedButton.icon(
                      onPressed: _loadAccount,
                      icon: const Icon(Icons.refresh),
                      label: const Text('Coba lagi'),
                    ),
                  ],
                ),
              ),
            )
          : account == null
          ? const Center(child: Text('Account tidak tersedia'))
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(20),
                    child: Column(
                      children: [
                        _AccountAvatar(account: account),
                        const SizedBox(height: 12),
                        Text(
                          account.accountName.isEmpty
                              ? 'Account'
                              : account.accountName,
                          style: Theme.of(context).textTheme.titleLarge,
                          textAlign: TextAlign.center,
                        ),
                        if (account.companyName.isNotEmpty) ...[
                          const SizedBox(height: 4),
                          Text(account.companyName),
                        ],
                        const SizedBox(height: 8),
                        Chip(
                          label: Text(account.isActive ? 'Aktif' : 'Nonaktif'),
                          avatar: Icon(
                            account.isActive
                                ? Icons.check_circle_outline
                                : Icons.pause_circle_outline,
                            size: 18,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                _AccountDetail(
                  icon: Icons.email_outlined,
                  label: 'Email account',
                  value: account.accountEmail,
                ),
                _AccountDetail(
                  icon: Icons.phone_android_outlined,
                  label: 'Mobile phone',
                  value: account.mobilePhone,
                ),
                _AccountDetail(
                  icon: Icons.phone_outlined,
                  label: 'Telepon',
                  value: account.phone,
                ),
                _AccountDetail(
                  icon: Icons.location_on_outlined,
                  label: 'Alamat',
                  value: account.address,
                ),
              ],
            ),
    );
  }
}

class _AccountAvatar extends StatelessWidget {
  const _AccountAvatar({required this.account});

  final AccountProfileModel account;

  @override
  Widget build(BuildContext context) {
    final logoUrl = account.companyLogoUrl.trim();
    return CircleAvatar(
      radius: 40,
      backgroundImage: logoUrl.isEmpty ? null : NetworkImage(logoUrl),
      onBackgroundImageError: logoUrl.isEmpty ? null : (_, _) {},
      child: logoUrl.isEmpty
          ? const Icon(Icons.business_outlined, size: 38)
          : null,
    );
  }
}

class _AccountDetail extends StatelessWidget {
  const _AccountDetail({
    required this.icon,
    required this.label,
    required this.value,
  });

  final IconData icon;
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: ListTile(
        leading: Icon(icon),
        title: Text(label, style: Theme.of(context).textTheme.bodySmall),
        subtitle: Text(value.isEmpty ? '-' : value),
      ),
    );
  }
}
