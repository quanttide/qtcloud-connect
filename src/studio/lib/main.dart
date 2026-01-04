import 'package:flutter/material.dart';
import 'screens/whiteboard_screen.dart';

void main() {
  runApp(const QtCloudConnectApp());
}

class QtCloudConnectApp extends StatelessWidget {
  const QtCloudConnectApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '量潮沟通云',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
        useMaterial3: true,
      ),
      home: const WhiteboardScreen(),
    );
  }
}
