import 'dart:async';
import 'dart:convert';

import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_local_notifications/flutter_local_notifications.dart';

class NotificationService {
  static final instance = NotificationService();

  NotificationService({
    FirebaseMessaging? messaging,
    FlutterLocalNotificationsPlugin? localNotifications,
  }) : _messaging = messaging ?? FirebaseMessaging.instance,
       _localNotifications =
           localNotifications ?? FlutterLocalNotificationsPlugin();

  static const _channelId = 'mobile_app_notifications';
  static const _channelName = 'App notifications';

  final FirebaseMessaging _messaging;
  final FlutterLocalNotificationsPlugin _localNotifications;
  final AndroidNotificationChannel _channel = const AndroidNotificationChannel(
    _channelId,
    _channelName,
    description: 'Updates and alerts from Open POS',
    importance: Importance.max,
  );

  StreamSubscription<String>? _tokenSubscription;
  StreamSubscription<RemoteMessage>? _foregroundSubscription;
  StreamSubscription<RemoteMessage>? _openedSubscription;
  Future<void> Function(String token)? _onTokenChanged;
  ValueChanged<Map<String, dynamic>>? _onNotificationTap;
  bool _initialized = false;

  Future<AuthorizationStatus> requestPermission() async {
    final settings = await _messaging.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );
    return settings.authorizationStatus;
  }

  Future<String?> initialize({
    Future<void> Function(String token)? onTokenChanged,
    ValueChanged<Map<String, dynamic>>? onNotificationTap,
  }) async {
    _onTokenChanged = onTokenChanged ?? _onTokenChanged;
    _onNotificationTap = onNotificationTap ?? _onNotificationTap;
    if (_initialized) {
      final token = await _messaging.getToken();
      if (token != null) await _onTokenChanged?.call(token);
      return token;
    }

    await _localNotifications.initialize(
      settings: const InitializationSettings(
        android: AndroidInitializationSettings('@mipmap/ic_launcher'),
        iOS: DarwinInitializationSettings(),
      ),
      onDidReceiveNotificationResponse: (response) {
        final payload = response.payload;
        if (payload == null || payload.isEmpty) return;
        final decoded = jsonDecode(payload);
        if (decoded is Map) {
          _onNotificationTap?.call(Map<String, dynamic>.from(decoded));
        }
      },
    );

    final androidNotifications = _localNotifications
        .resolvePlatformSpecificImplementation<
          AndroidFlutterLocalNotificationsPlugin
        >();
    await androidNotifications?.createNotificationChannel(_channel);
    // Firebase Messaging handles the platform notification permission prompt.
    await requestPermission();

    _tokenSubscription = _messaging.onTokenRefresh.listen((token) {
      final handler = _onTokenChanged;
      if (handler != null) {
        unawaited(handler(token));
      }
    });
    _foregroundSubscription = FirebaseMessaging.onMessage.listen(
      _showForegroundNotification,
    );
    _openedSubscription = FirebaseMessaging.onMessageOpenedApp.listen(
      (message) => _onNotificationTap?.call(message.data),
    );

    final launchDetails = await _localNotifications
        .getNotificationAppLaunchDetails();
    final launchPayload = launchDetails?.notificationResponse?.payload;
    if (launchDetails?.didNotificationLaunchApp == true &&
        launchPayload != null &&
        launchPayload.isNotEmpty) {
      final decoded = jsonDecode(launchPayload);
      if (decoded is Map) {
        _onNotificationTap?.call(Map<String, dynamic>.from(decoded));
      }
    }

    final initialMessage = await _messaging.getInitialMessage();
    if (initialMessage != null) {
      _onNotificationTap?.call(initialMessage.data);
    }

    _initialized = true;
    return _messaging.getToken();
  }

  Future<void> _showForegroundNotification(RemoteMessage message) async {
    final notification = message.notification;
    if (notification == null) return;

    await _localNotifications.show(
      id: message.hashCode,
      title: notification.title,
      body: notification.body,
      notificationDetails: NotificationDetails(
        android: AndroidNotificationDetails(
          _channel.id,
          _channel.name,
          channelDescription: _channel.description,
          importance: Importance.max,
          priority: Priority.high,
          icon:
              message.notification?.android?.smallIcon ?? '@mipmap/ic_launcher',
        ),
        iOS: const DarwinNotificationDetails(),
      ),
      payload: jsonEncode(message.data),
    );
  }

  Future<void> dispose() async {
    await _tokenSubscription?.cancel();
    await _foregroundSubscription?.cancel();
    await _openedSubscription?.cancel();
    _tokenSubscription = null;
    _foregroundSubscription = null;
    _openedSubscription = null;
    _initialized = false;
  }
}
