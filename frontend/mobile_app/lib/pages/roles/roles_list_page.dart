import 'package:flutter/material.dart';
import 'package:mobile_app/models/role_model.dart';
import 'package:mobile_app/services/role_service.dart';

import 'role_edit_page.dart';

class RolesListPage extends StatefulWidget {
  const RolesListPage({super.key, this.service});

  final RoleService? service;

  @override
  State<RolesListPage> createState() => _RolesListPageState();
}

class _RolesListPageState extends State<RolesListPage> {
  late final RoleService _service = widget.service ?? RoleService();
  final _searchController = TextEditingController();
  List<RoleModel> _roles = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadRoles();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadRoles() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final roles = await _service.getRoles();
      final search = _searchController.text.trim().toLowerCase();
      if (!mounted) return;
      setState(() {
        _roles = search.isEmpty
            ? roles
            : roles
                  .where((role) => role.name.toLowerCase().contains(search))
                  .toList(growable: false);
        _loading = false;
      });
    } on RoleServiceException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _loading = false;
      });
    }
  }

  Future<void> _openEditor([int? roleId]) async {
    final changed = await Navigator.of(context).push<bool>(
      MaterialPageRoute(
        builder: (_) => RoleEditPage(roleId: roleId, service: _service),
      ),
    );
    if (changed == true) _loadRoles();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Daftar Role')),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _openEditor(),
        icon: const Icon(Icons.add_moderator_outlined),
        label: const Text('Tambah Role'),
      ),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
            child: TextField(
              controller: _searchController,
              textInputAction: TextInputAction.search,
              onSubmitted: (_) => _loadRoles(),
              decoration: InputDecoration(
                hintText: 'Cari nama role',
                prefixIcon: const Icon(Icons.search),
                suffixIcon: IconButton(
                  onPressed: _loadRoles,
                  icon: const Icon(Icons.arrow_forward),
                ),
                border: const OutlineInputBorder(),
              ),
            ),
          ),
          Expanded(child: _buildContent()),
        ],
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
                onPressed: _loadRoles,
                icon: const Icon(Icons.refresh),
                label: const Text('Coba lagi'),
              ),
            ],
          ),
        ),
      );
    }
    if (_roles.isEmpty) {
      return RefreshIndicator(
        onRefresh: _loadRoles,
        child: ListView(
          children: const [
            SizedBox(height: 140),
            Icon(
              Icons.admin_panel_settings_outlined,
              size: 52,
              color: Colors.grey,
            ),
            SizedBox(height: 12),
            Center(child: Text('Role belum tersedia')),
          ],
        ),
      );
    }
    return RefreshIndicator(
      onRefresh: _loadRoles,
      child: ListView.separated(
        padding: const EdgeInsets.fromLTRB(16, 8, 16, 96),
        itemCount: _roles.length,
        separatorBuilder: (_, _) => const SizedBox(height: 8),
        itemBuilder: (context, index) {
          final role = _roles[index];
          return Card(
            margin: EdgeInsets.zero,
            child: ListTile(
              onTap: () => _openEditor(role.id),
              leading: const CircleAvatar(
                child: Icon(Icons.admin_panel_settings_outlined),
              ),
              title: Text(role.name),
              subtitle: Text('${role.actions.length} action'),
              trailing: const Icon(Icons.chevron_right),
            ),
          );
        },
      ),
    );
  }
}
