class NotificationModel {
  const NotificationModel({
    required this.id,
    required this.title,
    required this.message,
    required this.isRead,
    this.entityType = '',
    this.entityId = 0,
    this.entityNumber = '',
    this.action = '',
    this.readAt,
    this.createdAt,
  });

  final int id;
  final String title;
  final String message;
  final bool isRead;
  final String entityType;
  final int entityId;
  final String entityNumber;
  final String action;
  final DateTime? readAt;
  final DateTime? createdAt;

  NotificationModel copyWith({bool? isRead}) => NotificationModel(
    id: id,
    title: title,
    message: message,
    isRead: isRead ?? this.isRead,
    entityType: entityType,
    entityId: entityId,
    entityNumber: entityNumber,
    action: action,
    readAt: readAt,
    createdAt: createdAt,
  );

  factory NotificationModel.fromJson(Map<String, dynamic> json) {
    return NotificationModel(
      id: _asInt(json['id']),
      title: json['title']?.toString() ?? '',
      message: json['message']?.toString() ?? '',
      isRead: json['isRead'] == true,
      entityType: json['entityType']?.toString() ?? '',
      entityId: _asInt(json['entityId']),
      entityNumber: json['entityNumber']?.toString() ?? '',
      action: json['action']?.toString() ?? '',
      readAt: DateTime.tryParse(json['readAt']?.toString() ?? ''),
      createdAt: DateTime.tryParse(json['createdAt']?.toString() ?? ''),
    );
  }

  static int _asInt(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    return int.tryParse(value?.toString() ?? '') ?? 0;
  }
}
