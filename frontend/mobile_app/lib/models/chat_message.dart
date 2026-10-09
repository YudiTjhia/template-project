class ChatMessage {
  const ChatMessage({required this.text, required this.isUser, this.sentAt});

  final String text;
  final bool isUser;
  final DateTime? sentAt;
}
