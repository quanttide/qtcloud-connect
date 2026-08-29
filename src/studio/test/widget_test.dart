import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_connect_studio/main.dart';
import 'package:qtcloud_connect_studio/models/consensus.dart';
import 'package:qtcloud_connect_studio/screens/consensus_traceability_screen.dart';
import 'package:qtcloud_connect_studio/services/consensus_api.dart';

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
    expect(find.text('图谱概览'), findsOneWidget);
  });

  testWidgets('consensus screen exposes graph editing controls', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: ConsensusTraceabilityScreen(
          loadConsensuses: () async => const [
            Consensus(
              id: 'c1',
              title: '可编辑共识',
              description: '支持扩展链路。',
              createdAt: '2026-08-29T10:00:00Z',
              updatedAt: '2026-08-29T10:00:00Z',
            ),
          ],
        ),
      ),
    );

    await tester.pumpAndSettle();

    expect(find.byTooltip('添加共识'), findsOneWidget);
    expect(find.byTooltip('切换图谱'), findsOneWidget);
    expect(find.byTooltip('新建图谱'), findsOneWidget);
    expect(find.byTooltip('纳入已有共识'), findsOneWidget);
    expect(find.byTooltip('编辑共识'), findsOneWidget);
    expect(find.byTooltip('添加关联'), findsOneWidget);
  });

  testWidgets('user switches between independent consensus graphs', (
    tester,
  ) async {
    const paymentGraph = ConsensusGraph(
      id: 'payment-timeout',
      name: '支付超时演进',
      description: '支付超时问题的共识进展。',
      nodes: [
        Consensus(
          id: 'payment-node',
          title: '评估两个解决方案',
          description: '确认支付超时的处理路径。',
          createdAt: '2026-08-29T10:00:00Z',
          updatedAt: '2026-08-29T10:00:00Z',
        ),
      ],
      edges: [],
      createdAt: '2026-08-29T10:00:00Z',
      updatedAt: '2026-08-29T10:00:00Z',
    );
    const traceabilityGraph = ConsensusGraph(
      id: 'consensus-traceability',
      name: '共识追溯原型',
      description: '原始原型的可编辑参考图。',
      nodes: [
        Consensus(
          id: 'traceability-node',
          title: '发现支付超时问题',
          description: '原始共识链路的起点。',
          createdAt: '2026-08-29T10:00:00Z',
          updatedAt: '2026-08-29T10:00:00Z',
        ),
      ],
      edges: [],
      createdAt: '2026-08-29T10:00:00Z',
      updatedAt: '2026-08-29T10:00:00Z',
    );

    await tester.pumpWidget(
      MaterialApp(
        home: ConsensusTraceabilityScreen(
          apiClient: _GraphApiClient(const [paymentGraph, traceabilityGraph]),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('支付超时演进'), findsOneWidget);

    await tester.tap(find.byTooltip('切换图谱'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('共识追溯原型'));
    await tester.pumpAndSettle();

    expect(find.text('共识追溯原型'), findsOneWidget);
    expect(find.text('发现支付超时问题'), findsOneWidget);
  });
}

class _GraphApiClient extends ConsensusApiClient {
  const _GraphApiClient(this.graphs);

  final List<ConsensusGraph> graphs;

  @override
  Future<List<ConsensusGraph>> listGraphs() async => graphs;

  @override
  Future<ConsensusGraph> getGraph(String id) async =>
      graphs.singleWhere((graph) => graph.id == id);
}
