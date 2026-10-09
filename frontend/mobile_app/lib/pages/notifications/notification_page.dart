import 'package:flutter/material.dart';
import 'package:mobile_app/models/notification_model.dart';
import 'package:mobile_app/services/notification_api_service.dart';

class NotificationPage extends StatefulWidget {
  const NotificationPage({super.key, this.service});

  final NotificationApiService? service;

  @override
  State<NotificationPage> createState() => _NotificationPageState();
}

class _NotificationPageState extends State<NotificationPage> {
  late final NotificationApiService _service =
      widget.service ?? NotificationApiService();
  List<NotificationModel> _notifications = const [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadNotifications();
  }

  Future<void> _loadNotifications() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final items = await _service.getNotifications();
      if (!mounted) return;
      setState(() {
        _notifications = items;
        _loading = false;
      });
    } on NotificationApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _loading = false;
      });
    }
  }

  Future<void> _openNotification(NotificationModel notification) async {
    if (!notification.isRead) {
      try {
        await _service.markAsRead(notification.id);
        if (!mounted) return;
        setState(() {
          _notifications = _notifications
              .map(
                (item) => item.id == notification.id
                    ? item.copyWith(isRead: true)
                    : item,
              )
              .toList(growable: false);
        });
      } on NotificationApiException catch (error) {
        if (!mounted) return;
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text(error.message)));
      }
    }
    if (!mounted) return;
    await showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.notifications_active_outlined),
        title: Text(notification.title),
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(notification.message),
              if (notification.entityNumber.isNotEmpty) ...[
                const SizedBox(height: 16),
                Text(
                  'Nomor: ${notification.entityNumber}',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
              if (notification.createdAt != null) ...[
                const SizedBox(height: 8),
                Text(
                  _formatDate(notification.createdAt!),
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ],
          ),
        ),
        actions: [
          FilledButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Tutup'),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    final local = date.toLocal();
    final day = local.day.toString().padLeft(2, '0');
    final month = local.month.toString().padLeft(2, '0');
    final hour = local.hour.toString().padLeft(2, '0');
    final minute = local.minute.toString().padLeft(2, '0');
    return '$day/$month/${local.year} $hour:$minute';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Notifikasi')),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_loading) return const Center(child: CircularProgressIndicator());
    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.wifi_off_outlined, size: 42),
              const SizedBox(height: 12),
              Text(_error!, textAlign: TextAlign.center),
              const SizedBox(height: 12),
              OutlinedButton.icon(
                onPressed: _loadNotifications,
                icon: const Icon(Icons.refresh),
                label: const Text('Coba lagi'),
              ),
            ],
          ),
        ),
      );
    }
    if (_notifications.isEmpty) {
      return RefreshIndicator(
        onRefresh: _loadNotifications,
        child: ListView(
          children: const [
            SizedBox(height: 180),
            Icon(Icons.notifications_none, size: 52, color: Colors.grey),
            SizedBox(height: 12),
            Center(child: Text('Belum ada notifikasi')),
          ],
        ),
      );
    }
    return RefreshIndicator(
      onRefresh: _loadNotifications,
      child: ListView.separated(
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: _notifications.length,
        separatorBuilder: (_, _) => const Divider(height: 1, indent: 72),
        itemBuilder: (context, index) {
          final notification = _notifications[index];
          return ListTile(
            onTap: () => _openNotification(notification),
            leading: CircleAvatar(
              backgroundColor: notification.isRead
                  ? Colors.grey.shade100
                  : Theme.of(context).colorScheme.primaryContainer,
              child: Icon(
                notification.isRead
                    ? Icons.notifications_none
                    : Icons.notifications_active_outlined,
                color: notification.isRead
                    ? Colors.grey
                    : Theme.of(context).colorScheme.primary,
              ),
            ),
            title: Text(
              notification.title,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(
                fontWeight: notification.isRead
                    ? FontWeight.w500
                    : FontWeight.w700,
              ),
            ),
            subtitle: Text(
              notification.message,
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            trailing: notification.isRead
                ? null
                : Icon(
                    Icons.circle,
                    size: 10,
                    color: Theme.of(context).colorScheme.primary,
                  ),
          );
        },
      ),
    );
  }
}
