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

  testWidgets(
    'dragging a node persists its graph position after switching graphs',
    (tester) async {
      const graph = ConsensusGraph(
        id: 'drag-graph',
        name: '拖拽持久化',
        description: '节点坐标需要在切换图谱后保留。',
        nodes: [
          Consensus(
            id: 'drag-node',
            title: '可拖动节点',
            description: '拖动结束后保存坐标。',
            createdAt: '2026-08-29T10:00:00Z',
            updatedAt: '2026-08-29T10:00:00Z',
          ),
        ],
        edges: [],
        createdAt: '2026-08-29T10:00:00Z',
        updatedAt: '2026-08-29T10:00:00Z',
      );
      const otherGraph = ConsensusGraph(
        id: 'other-graph',
        name: '另一张图',
        description: '切换后再切回来。',
        nodes: [
          Consensus(
            id: 'other-node',
            title: '另一张图节点',
            description: '独立图谱。',
            createdAt: '2026-08-29T10:00:00Z',
            updatedAt: '2026-08-29T10:00:00Z',
          ),
        ],
        edges: [],
        createdAt: '2026-08-29T10:00:00Z',
        updatedAt: '2026-08-29T10:00:00Z',
      );
      final api = _GraphApiClient(const [graph, otherGraph]);

      await tester.pumpWidget(
        MaterialApp(home: ConsensusTraceabilityScreen(apiClient: api)),
      );
      await tester.pumpAndSettle();

      final initialPosition = tester.getTopLeft(find.text('可拖动节点'));
      final gesture = await tester.startGesture(
        tester.getCenter(find.text('可拖动节点')),
      );
      await gesture.moveBy(const Offset(24, 24));
      await gesture.moveBy(const Offset(96, 40));
      await gesture.up();
      await tester.pumpAndSettle();

      expect(api.positionUpdates, hasLength(1));
      expect(api.positionUpdates.single.graphId, 'drag-graph');
      expect(api.positionUpdates.single.consensusId, 'drag-node');
      expect(api.positionUpdates.single.x, greaterThan(72));
      expect(api.positionUpdates.single.y, greaterThan(72));

      await tester.tap(find.byTooltip('切换图谱'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('另一张图'));
      await tester.pumpAndSettle();
      await tester.tap(find.byTooltip('切换图谱'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('拖拽持久化'));
      await tester.pumpAndSettle();

      expect(find.text('可拖动节点'), findsOneWidget);
      expect(
        tester.getTopLeft(find.text('可拖动节点')).dx,
        greaterThan(initialPosition.dx),
      );
      expect(
        tester.getTopLeft(find.text('可拖动节点')).dy,
        greaterThan(initialPosition.dy),
      );
    },
  );

  testWidgets('failed position save restores the previous visible position', (
    tester,
  ) async {
    const graph = ConsensusGraph(
      id: 'failed-save-graph',
      name: '保存失败',
      description: '失败后回到服务端位置。',
      nodes: [
        Consensus(
          id: 'failed-save-node',
          title: '回滚节点',
          description: '保存失败后不保留错误坐标。',
          createdAt: '2026-08-29T10:00:00Z',
          updatedAt: '2026-08-29T10:00:00Z',
        ),
      ],
      edges: [],
      createdAt: '2026-08-29T10:00:00Z',
      updatedAt: '2026-08-29T10:00:00Z',
    );
    final api = _FailingGraphApiClient([graph]);

    await tester.pumpWidget(
      MaterialApp(home: ConsensusTraceabilityScreen(apiClient: api)),
    );
    await tester.pumpAndSettle();

    final initialPosition = tester.getTopLeft(find.text('回滚节点'));
    final gesture = await tester.startGesture(
      tester.getCenter(find.text('回滚节点')),
    );
    await gesture.moveBy(const Offset(24, 24));
    await gesture.moveBy(const Offset(96, 40));
    await gesture.up();
    await tester.pumpAndSettle();

    expect(
      tester.getTopLeft(find.text('回滚节点')).dx,
      closeTo(initialPosition.dx, 0.1),
    );
    expect(
      tester.getTopLeft(find.text('回滚节点')).dy,
      closeTo(initialPosition.dy, 0.1),
    );
  });

  testWidgets('tapping a node does not persist its graph position', (
    tester,
  ) async {
    const graph = ConsensusGraph(
      id: 'tap-graph',
      name: '点击节点',
      description: '点击只查看详情。',
      nodes: [
        Consensus(
          id: 'tap-node',
          title: '查看详情',
          description: '不应产生坐标保存请求。',
          createdAt: '2026-08-29T10:00:00Z',
          updatedAt: '2026-08-29T10:00:00Z',
        ),
      ],
      edges: [],
      createdAt: '2026-08-29T10:00:00Z',
      updatedAt: '2026-08-29T10:00:00Z',
    );
    final api = _GraphApiClient(const [graph]);

    await tester.pumpWidget(
      MaterialApp(home: ConsensusTraceabilityScreen(apiClient: api)),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.text('查看详情'));
    await tester.pumpAndSettle();

    expect(api.positionUpdates, isEmpty);
  });
}

class _FailingGraphApiClient extends _GraphApiClient {
  _FailingGraphApiClient(super.graphs);

  @override
  Future<ConsensusGraph> updateNodePosition({
    required String graphId,
    required String consensusId,
    required double x,
    required double y,
  }) async {
    throw const ConsensusApiException('simulated position save failure');
  }
}

class _GraphApiClient extends ConsensusApiClient {
  _GraphApiClient(List<ConsensusGraph> graphs) : graphs = [...graphs];

  final List<ConsensusGraph> graphs;
  final List<_NodePositionUpdate> positionUpdates = [];

  @override
  Future<List<ConsensusGraph>> listGraphs() async => graphs;

  @override
  Future<ConsensusGraph> getGraph(String id) async =>
      graphs.singleWhere((graph) => graph.id == id);

  @override
  Future<ConsensusGraph> updateNodePosition({
    required String graphId,
    required String consensusId,
    required double x,
    required double y,
  }) async {
    positionUpdates.add(
      _NodePositionUpdate(
        graphId: graphId,
        consensusId: consensusId,
        x: x,
        y: y,
      ),
    );
    final graph = graphs.singleWhere((graph) => graph.id == graphId);
    final updated = graph.copyWith(
      nodePositions: {
        ...graph.nodePositions,
        consensusId: ConsensusGraphNodePosition(x: x, y: y),
      },
    );
    graphs[graphs.indexOf(graph)] = updated;
    return updated;
  }
}

class _NodePositionUpdate {
  const _NodePositionUpdate({
    required this.graphId,
    required this.consensusId,
    required this.x,
    required this.y,
  });

  final String graphId;
  final String consensusId;
  final double x;
  final double y;
}
