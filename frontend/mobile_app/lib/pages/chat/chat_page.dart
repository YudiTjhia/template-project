import 'package:flutter/material.dart';

import '../../models/chat_message.dart';
import '../../navigation/app_bottom_menu.dart';
import '../../navigation/app_drawer_menu.dart';
import '../../shared/widgets/bottom_menu.dart';
import '../../widgets/chat_component.dart';

class ChatPage extends StatefulWidget {
  const ChatPage({
    super.key,
    this.enableDrawer = true,
    this.useBottomMenu = true,
  });

  final bool enableDrawer;
  final bool useBottomMenu;

  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  final _messageController = TextEditingController();
  final List<ChatMessage> _messages = [
    ChatMessage(
      text: 'Halo! Ada yang bisa saya bantu?',
      isUser: false,
      sentAt: DateTime.now(),
    ),
  ];

  @override
  void dispose() {
    _messageController.dispose();
    super.dispose();
  }

  void _sendMessage() {
    final text = _messageController.text.trim();
    if (text.isEmpty) return;
    setState(() {
      _messages.add(
        ChatMessage(text: text, isUser: true, sentAt: DateTime.now()),
      );
      _messageController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('AI Chat')),
      drawer: widget.enableDrawer
          ? const AppDrawerMenu(selectedRoute: '/ai-chat')
          : null,
      body: ChatComponent(
        messages: _messages,
        controller: _messageController,
        onSend: _sendMessage,
      ),
      bottomNavigationBar: widget.useBottomMenu
          ? const AppBottomMenu(selected: MenuDestination.aiChat)
          : null,
    );
  }
}
