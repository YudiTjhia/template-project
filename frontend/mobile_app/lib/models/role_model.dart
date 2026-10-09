class RoleModel {
  const RoleModel({
    required this.id,
    required this.name,
    required this.actions,
  });

  final int id;
  final String name;
  final List<String> actions;

  factory RoleModel.fromJson(Map<String, dynamic> json) => RoleModel(
    id: _asInt(json['id']),
    name: json['role_name']?.toString() ?? '',
    actions: (json['actions'] as List? ?? const [])
        .map((action) => action.toString())
        .toList(growable: false),
  );

  Map<String, dynamic> toJson() => {
    'role_name': name.trim(),
    'actions': actions,
  };

  static int _asInt(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    return int.tryParse(value?.toString() ?? '') ?? 0;
  }
}
