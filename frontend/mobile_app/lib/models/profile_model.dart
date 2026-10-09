class ProfileModel {
  const ProfileModel({
    required this.name,
    required this.username,
    required this.email,
    this.photoUrl,
  });

  final String name;
  final String username;
  final String email;
  final String? photoUrl;

  factory ProfileModel.fromJson(Map<String, dynamic> json) {
    final username = json['username']?.toString() ?? '';
    final email =
        json['user_email']?.toString() ?? json['email']?.toString() ?? '';
    final name = json['name']?.toString().trim();
    final photo = json['photo_url'] ?? json['photoUrl'] ?? json['avatar'];

    return ProfileModel(
      name: name == null || name.isEmpty ? username : name,
      username: username,
      email: email,
      photoUrl: photo?.toString(),
    );
  }
}
