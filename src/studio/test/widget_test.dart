import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_connect_studio/main.dart';
import 'package:qtcloud_connect_studio/models/consensus.dart';
import 'package:qtcloud_connect_studio/screens/consensus_traceability_screen.dart';

void main() {
  testWidgets('app opens on consensus traceability surface', (tester) async {
    await tester.pumpWidget(const QtCloudConnectApp());

    expect(find.text('共识追溯图'), findsOneWidget);
    expect(find.text('消息'), findsOneWidget);
    expect(find.text('备忘'), findsOneWidget);
  });

  testWidgets('consensus screen renders provider data', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        home: ConsensusTraceabilityScreen(
          loadConsensuses: () async => const [
            Consensus(
              id: 'c1',
              title: 'CLI 可以记录真实共识',
              description: 'Provider 持久化后 Studio 展示。',
              createdAt: '2026-08-29T10:00:00Z',
              updatedAt: '2026-08-29T10:00:00Z',
            ),
          ],
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.text('CLI 可以记录真实共识'), findsOneWidget);
    expect(find.text('Provider 持久化后 Studio 展示。'), findsOneWidget);
    expect(find.text('共识总数'), findsOneWidget);
  });
}
