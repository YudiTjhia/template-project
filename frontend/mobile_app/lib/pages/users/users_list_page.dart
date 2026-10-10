import 'package:flutter/material.dart';

import 'package:mobile_app/models/user_model.dart';
import 'package:mobile_app/services/user_service.dart';

import '../../navigation/app_drawer_menu.dart';
import '../../shared/widgets/responsive_page_layout.dart';

import 'user_edit_page.dart';

class UsersListPage extends StatefulWidget {
  const UsersListPage({super.key, this.service, this.enableDrawer = true});

  final UserService? service;
  final bool enableDrawer;

  @override
  State<UsersListPage> createState() => _UsersListPageState();
}

class _UsersListPageState extends State<UsersListPage> {
  late final UserService _service = widget.service ?? UserService();
  final _searchController = TextEditingController();
  List<UserModel> _users = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadUsers();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadUsers() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final users = await _service.getUsers(search: _searchController.text);
      if (!mounted) return;
      setState(() {
        _users = users;
        _loading = false;
      });
    } on UserServiceException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _loading = false;
      });
    }
  }

  Future<void> _openEditor([int? userId]) async {
    final changed = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => UserEditPage(userId: userId, service: _service),
      ),
    );
    if (changed == true) _loadUsers();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Daftar User')),
      drawer: widget.enableDrawer
          ? const AppDrawerMenu(selectedRoute: '/users')
          : null,
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _openEditor(),
        icon: const Icon(Icons.person_add_alt_1),
        label: const Text('Tambah User'),
      ),
      body: ResponsivePageLayout(
        maxWidth: 1100,
        compactPadding: 0,
        mediumPadding: 0,
        expandedPadding: 0,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
              child: TextField(
                controller: _searchController,
                textInputAction: TextInputAction.search,
                onSubmitted: (_) => _loadUsers(),
                decoration: InputDecoration(
                  hintText: 'Cari nama, email, atau nomor telepon',
                  prefixIcon: const Icon(Icons.search),
                  suffixIcon: IconButton(
                    tooltip: 'Cari',
                    onPressed: _loadUsers,
                    icon: const Icon(Icons.arrow_forward),
                  ),
                  border: const OutlineInputBorder(),
                ),
              ),
            ),
            Expanded(child: _buildContent()),
          ],
        ),
      ),
    );
  }

  Widget _buildContent() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(_error!, textAlign: TextAlign.center),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: _loadUsers,
                icon: const Icon(Icons.refresh),
                label: const Text('Coba lagi'),
              ),
            ],
          ),
        ),
      );
    }
    if (_users.isEmpty) {
      return RefreshIndicator(
        onRefresh: _loadUsers,
        child: ListView(
          children: const [
            SizedBox(height: 140),
            Icon(Icons.people_outline, size: 52, color: Colors.grey),
            SizedBox(height: 12),
            Center(child: Text('User belum tersedia')),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadUsers,
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 96),
        itemCount: _users.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final user = _users[index];
          return Card(
            margin: EdgeInsets.zero,
            child: ListTile(
              onTap: () => _openEditor(user.id),
              leading: CircleAvatar(
                child: Text(
                  user.username.isEmpty
                      ? '?'
                      : user.username.characters.first.toUpperCase(),
                ),
              ),
              title: Text(user.username),
              subtitle: Text(
                [
                  user.email,
                  if (user.userRoles.isNotEmpty) user.userRoles.join(', '),
                ].join('\n'),
              ),
              isThreeLine: user.userRoles.isNotEmpty,
              trailing: Icon(
                user.isActive ? Icons.check_circle : Icons.block,
                color: user.isActive ? Colors.green : Colors.grey,
              ),
            ),
          );
        },
      ),
    );
  }
}
