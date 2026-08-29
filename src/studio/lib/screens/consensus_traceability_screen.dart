import 'dart:math' as math;

import 'package:flutter/material.dart';

import '../models/consensus.dart';
import '../services/consensus_api.dart';

typedef ConsensusLoader = Future<List<Consensus>> Function();
typedef ConsensusGraphLoader = Future<ConsensusGraph> Function();

class ConsensusTraceabilityScreen extends StatefulWidget {
  const ConsensusTraceabilityScreen({
    super.key,
    this.loadConsensuses,
    this.loadGraph,
    this.apiClient = const ConsensusApiClient(),
  });

  final ConsensusLoader? loadConsensuses;
  final ConsensusGraphLoader? loadGraph;
  final ConsensusApiClient apiClient;

  @override
  State<ConsensusTraceabilityScreen> createState() =>
      _ConsensusTraceabilityScreenState();
}

class _ConsensusTraceabilityScreenState
    extends State<ConsensusTraceabilityScreen> {
  late Future<ConsensusGraph> _graphFuture;
  ConsensusGraph? _graph;
  List<ConsensusGraph> _graphs = const [];
  String? _selectedGraphId;
  int _graphRequestVersion = 0;
  String? _selectedId;
  Object? _error;
  bool _isSaving = false;
  final Map<String, Offset> _nodePositions = {};

  bool get _isLocal =>
      widget.loadConsensuses != null || widget.loadGraph != null;

  @override
  void initState() {
    super.initState();
    _graphFuture = _loadInitialGraph();
    _cacheLoadedGraph(_graphFuture);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Row(
          children: [
            NavigationRail(
              selectedIndex: 1,
              labelType: NavigationRailLabelType.all,
              destinations: const [
                NavigationRailDestination(
                  icon: Icon(Icons.chat_bubble_outline),
                  label: Text('消息'),
                ),
                NavigationRailDestination(
                  icon: Icon(Icons.account_tree_outlined),
                  selectedIcon: Icon(Icons.account_tree),
                  label: Text('共识'),
                ),
                NavigationRailDestination(
                  icon: Icon(Icons.description_outlined),
                  label: Text('备忘'),
                ),
              ],
            ),
            const VerticalDivider(width: 1),
            Expanded(
              child: FutureBuilder<ConsensusGraph>(
                future: _graphFuture,
                builder: (context, snapshot) {
                  final graph = _graph ?? snapshot.data;
                  return _ConsensusSurface(
                    graph: graph,
                    selectedId: _selectedId,
                    isLoading:
                        snapshot.connectionState == ConnectionState.waiting,
                    error:
                        _error ?? (snapshot.hasError ? snapshot.error : null),
                    isSaving: _isSaving,
                    nodePositions: _nodePositions,
                    canAddExistingConsensus: !_isLocal,
                    graphs: _graphs,
                    selectedGraphId: _selectedGraphId,
                    canManageGraphs: !_isLocal,
                    onRefresh: _refresh,
                    onSelectGraph: _selectGraph,
                    onCreateGraph: _createGraph,
                    onSelect: (id) => setState(() => _selectedId = id),
                    onMoveNode: _setNodePosition,
                    onAddConsensus: _addConsensus,
                    onAddExistingConsensus: _addExistingConsensus,
                    onEditConsensus: _editSelectedConsensus,
                    onAddRelation: _addRelation,
                    onRemoveNode: _removeSelectedNode,
                    onRemoveEdge: _removeEdge,
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Future<ConsensusGraph> _loadInitialGraph() async {
    if (widget.loadGraph != null) {
      return widget.loadGraph!();
    }
    if (widget.loadConsensuses != null) {
      return _localGraph(nodes: await widget.loadConsensuses!());
    }

    final graphs = await widget.apiClient.listGraphs();
    _graphs = graphs;
    if (graphs.isNotEmpty) {
      final selected = graphs
          .where((graph) => graph.id == _selectedGraphId)
          .firstOrNull;
      final graph = selected ?? graphs.first;
      _selectedGraphId = graph.id;
      return widget.apiClient.getGraph(graph.id);
    }

    final graph = await widget.apiClient.createGraph(
      name: '共识追溯图',
      description: '团队共识的可编辑决策网络。',
    );
    _graphs = [graph];
    _selectedGraphId = graph.id;
    final consensuses = await widget.apiClient.listConsensuses();
    var current = graph;
    for (final consensus in consensuses) {
      current = await widget.apiClient.addNode(
        graphId: current.id,
        consensusId: consensus.id,
      );
    }
    return current;
  }

  Future<void> _createGraph() async {
    if (_isLocal || _isSaving) {
      return;
    }
    final draft = await _showGraphDialog(context);
    if (draft == null) {
      return;
    }

    setState(() {
      _isSaving = true;
      _error = null;
    });
    try {
      final graph = await widget.apiClient.createGraph(
        name: draft.name,
        description: draft.description,
      );
      if (!mounted) {
        return;
      }
      setState(() {
        _graphs = [..._graphs, graph];
        _selectedGraphId = graph.id;
        _graphRequestVersion++;
        _selectedId = null;
        _graph = graph;
        _graphFuture = Future.value(graph);
        _nodePositions.clear();
        _isSaving = false;
      });
    } catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isSaving = false;
        _error = error;
      });
      _showMessage(_formatError(error));
    }
  }

  void _selectGraph(String graphId) {
    if (_isLocal || _isSaving || graphId == _selectedGraphId) {
      return;
    }
    final future = widget.apiClient.getGraph(graphId);
    setState(() {
      _selectedGraphId = graphId;
      _graphRequestVersion++;
      _selectedId = null;
      _graph = null;
      _error = null;
      _nodePositions.clear();
      _graphFuture = future;
      _cacheLoadedGraph(future);
    });
  }

  ConsensusGraph _localGraph({
    List<Consensus> nodes = const [],
    List<ConsensusRelation> edges = const [],
  }) {
    final now = DateTime.now().toUtc().toIso8601String();
    return ConsensusGraph(
      id: 'local-graph',
      name: '共识追溯图',
      description: '可编辑的共识决策网络。',
      nodes: List<Consensus>.unmodifiable(nodes),
      edges: List<ConsensusRelation>.unmodifiable(edges),
      createdAt: now,
      updatedAt: now,
    );
  }

  void _cacheLoadedGraph(Future<ConsensusGraph> future) {
    final requestVersion = _graphRequestVersion;
    future
        .then<void>((graph) {
          if (mounted && requestVersion == _graphRequestVersion) {
            setState(() => _graph = graph);
          }
        })
        .catchError((_) {});
  }

  void _setNodePosition(String id, Offset position) {
    setState(() {
      _nodePositions[id] = position;
    });
  }

  Future<void> _addConsensus() async {
    final draft = await _showConsensusDialog(context);
    if (draft == null) {
      return;
    }
    await _runMutation(() async {
      final graph = _currentGraph;
      if (_isLocal) {
        final now = DateTime.now().toUtc().toIso8601String();
        final consensus = Consensus(
          id: 'local-${DateTime.now().microsecondsSinceEpoch}',
          title: draft.title,
          description: draft.description,
          createdAt: now,
          updatedAt: now,
        );
        return graph.copyWith(
          nodes: [...graph.nodes, consensus],
          updatedAt: now,
        );
      }
      final consensus = await widget.apiClient.createConsensus(
        title: draft.title,
        description: draft.description,
      );
      return widget.apiClient.addNode(
        graphId: graph.id,
        consensusId: consensus.id,
      );
    });
  }

  Future<void> _addExistingConsensus() async {
    try {
      final graph = await _safeGraph();
      if (!mounted || graph == null) {
        return;
      }
      final available = _isLocal
          ? const <Consensus>[]
          : (await widget.apiClient.listConsensuses())
                .where(
                  (consensus) =>
                      !graph.nodes.any((node) => node.id == consensus.id),
                )
                .toList(growable: false);
      if (!mounted) {
        return;
      }
      if (available.isEmpty) {
        _showMessage('没有可纳入当前图的已有共识。');
        return;
      }
      final consensusId = await _showExistingConsensusDialog(
        context,
        available,
      );
      if (consensusId == null) {
        return;
      }
      await _runMutation(
        () => widget.apiClient.addNode(
          graphId: graph.id,
          consensusId: consensusId,
        ),
      );
    } catch (error) {
      if (mounted) {
        setState(() => _error = error);
        _showMessage(_formatError(error));
      }
    }
  }

  Future<void> _editSelectedConsensus() async {
    final selected = _selectedConsensus;
    if (selected == null || selected.status != 'proposed') {
      return;
    }
    final draft = await _showConsensusDialog(context, initial: selected);
    if (draft == null) {
      return;
    }
    await _runMutation(() async {
      final graph = _currentGraph;
      final now = DateTime.now().toUtc().toIso8601String();
      if (_isLocal) {
        return graph.copyWith(
          nodes: [
            for (final node in graph.nodes)
              node.id == selected.id
                  ? node.copyWith(
                      title: draft.title,
                      description: draft.description,
                      updatedAt: now,
                    )
                  : node,
          ],
          updatedAt: now,
        );
      }
      final updated = await widget.apiClient.updateConsensus(
        id: selected.id,
        title: draft.title,
        description: draft.description,
      );
      return graph.copyWith(
        nodes: [
          for (final node in graph.nodes)
            node.id == updated.id ? updated : node,
        ],
        updatedAt: now,
      );
    });
  }

  Future<void> _addRelation() async {
    final graph = _graph ?? await _safeGraph();
    if (!mounted || graph == null || graph.nodes.length < 2) {
      _showMessage('至少需要两个共识节点才能建立关联。');
      return;
    }
    final draft = await _showRelationDialog(context, graph.nodes);
    if (draft == null) {
      return;
    }
    await _runMutation(() async {
      final current = _currentGraph;
      if (_isLocal) {
        final now = DateTime.now().toUtc().toIso8601String();
        final relation = ConsensusRelation(
          id: 'local-relation-${DateTime.now().microsecondsSinceEpoch}',
          from: draft.from,
          to: draft.to,
          relationType: draft.relationType,
        );
        if (_createsCycle(current, relation)) {
          throw const ConsensusApiException('edge would create a cycle');
        }
        return current.copyWith(
          edges: [...current.edges, relation],
          updatedAt: now,
        );
      }
      return widget.apiClient.createGraphRelation(
        graphId: current.id,
        from: draft.from,
        to: draft.to,
        relationType: draft.relationType,
      );
    });
  }

  Future<void> _removeSelectedNode() async {
    final selected = _selectedConsensus;
    if (selected == null) {
      return;
    }
    if (!await _confirm(
      title: '移出当前图',
      message: '只从当前图移除“${selected.title}”，不会删除共识本身。',
    )) {
      return;
    }
    await _runMutation(() async {
      final graph = _currentGraph;
      if (_isLocal) {
        final now = DateTime.now().toUtc().toIso8601String();
        return graph.copyWith(
          nodes: graph.nodes.where((node) => node.id != selected.id).toList(),
          edges: graph.edges
              .where(
                (edge) => edge.from != selected.id && edge.to != selected.id,
              )
              .toList(),
          updatedAt: now,
        );
      }
      return widget.apiClient.removeNode(
        graphId: graph.id,
        consensusId: selected.id,
      );
    });
    if (mounted) {
      setState(() => _selectedId = null);
    }
  }

  Future<void> _removeEdge(ConsensusRelation edge) async {
    if (!await _confirm(
      title: '移除关联',
      message: '从当前图移除这条“${edge.relationType}”关联？',
    )) {
      return;
    }
    await _runMutation(() async {
      final graph = _currentGraph;
      if (_isLocal) {
        return graph.copyWith(
          edges: graph.edges.where((item) => item.id != edge.id).toList(),
          updatedAt: DateTime.now().toUtc().toIso8601String(),
        );
      }
      return widget.apiClient.removeEdge(
        graphId: graph.id,
        relationId: edge.id,
      );
    });
  }

  Future<void> _runMutation(Future<ConsensusGraph> Function() mutation) async {
    setState(() {
      _isSaving = true;
      _error = null;
    });
    try {
      final graph = await mutation();
      if (!mounted) {
        return;
      }
      setState(() {
        _isSaving = false;
        if (_isLocal || graph.id == _selectedGraphId) {
          _graph = graph;
        }
      });
    } catch (error) {
      if (!mounted) {
        return;
      }
      setState(() {
        _isSaving = false;
        _error = error;
      });
      _showMessage(_formatError(error));
    }
  }

  Future<ConsensusGraph?> _safeGraph() async {
    try {
      return _graph ??= await _graphFuture;
    } catch (error) {
      if (mounted) {
        setState(() => _error = error);
      }
      return null;
    }
  }

  bool _createsCycle(ConsensusGraph graph, ConsensusRelation candidate) {
    final nodeIDs = graph.nodes.map((node) => node.id).toSet();
    if (candidate.from == candidate.to ||
        !nodeIDs.contains(candidate.from) ||
        !nodeIDs.contains(candidate.to)) {
      return true;
    }

    final outgoing = {for (final nodeID in nodeIDs) nodeID: <String>[]};
    final incomingCount = {for (final nodeID in nodeIDs) nodeID: 0};
    for (final edge in [...graph.edges, candidate]) {
      if (!nodeIDs.contains(edge.from) || !nodeIDs.contains(edge.to)) {
        return true;
      }
      outgoing[edge.from]!.add(edge.to);
      incomingCount[edge.to] = incomingCount[edge.to]! + 1;
    }

    final queue = incomingCount.entries
        .where((entry) => entry.value == 0)
        .map((entry) => entry.key)
        .toList();
    var visited = 0;
    for (var index = 0; index < queue.length; index++) {
      final current = queue[index];
      visited++;
      for (final next in outgoing[current]!) {
        incomingCount[next] = incomingCount[next]! - 1;
        if (incomingCount[next] == 0) {
          queue.add(next);
        }
      }
    }
    return visited != nodeIDs.length;
  }

  ConsensusGraph get _currentGraph {
    final graph = _graph;
    if (graph == null) {
      throw StateError('Graph is not loaded');
    }
    return graph;
  }

  Consensus? get _selectedConsensus {
    final id = _selectedId;
    if (id == null) {
      return null;
    }
    return (_graph?.nodes ?? const <Consensus>[])
        .where((node) => node.id == id)
        .firstOrNull;
  }

  void _refresh() {
    if (_isSaving) {
      return;
    }
    setState(() {
      _graph = null;
      _selectedId = null;
      _error = null;
      _nodePositions.clear();
      _graphRequestVersion++;
      _graphFuture = _loadInitialGraph();
      _cacheLoadedGraph(_graphFuture);
    });
  }

  Future<bool> _confirm({
    required String title,
    required String message,
  }) async {
    return (await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text(title),
            content: Text(message),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(context, false),
                child: const Text('取消'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(context, true),
                child: const Text('确定'),
              ),
            ],
          ),
        )) ??
        false;
  }

  void _showMessage(String message) {
    if (!mounted) {
      return;
    }
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  String _formatError(Object error) {
    if (error is ConsensusApiException) {
      return error.message;
    }
    return '操作失败，请检查 Provider 连接后重试。';
  }
}

class _ConsensusSurface extends StatelessWidget {
  const _ConsensusSurface({
    required this.graph,
    required this.selectedId,
    required this.isLoading,
    required this.error,
    required this.isSaving,
    required this.nodePositions,
    required this.canAddExistingConsensus,
    required this.graphs,
    required this.selectedGraphId,
    required this.canManageGraphs,
    required this.onRefresh,
    required this.onSelectGraph,
    required this.onCreateGraph,
    required this.onSelect,
    required this.onMoveNode,
    required this.onAddConsensus,
    required this.onAddExistingConsensus,
    required this.onEditConsensus,
    required this.onAddRelation,
    required this.onRemoveNode,
    required this.onRemoveEdge,
  });

  final ConsensusGraph? graph;
  final String? selectedId;
  final bool isLoading;
  final Object? error;
  final bool isSaving;
  final Map<String, Offset> nodePositions;
  final bool canAddExistingConsensus;
  final List<ConsensusGraph> graphs;
  final String? selectedGraphId;
  final bool canManageGraphs;
  final VoidCallback onRefresh;
  final ValueChanged<String> onSelectGraph;
  final VoidCallback onCreateGraph;
  final ValueChanged<String> onSelect;
  final void Function(String id, Offset position) onMoveNode;
  final VoidCallback onAddConsensus;
  final VoidCallback onAddExistingConsensus;
  final VoidCallback onEditConsensus;
  final VoidCallback onAddRelation;
  final VoidCallback onRemoveNode;
  final ValueChanged<ConsensusRelation> onRemoveEdge;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isCompact = constraints.maxWidth < 900;
        final graphPanel = _GraphPanel(
          graph: graph,
          selectedId: selectedId,
          isLoading: isLoading,
          error: error,
          isSaving: isSaving,
          nodePositions: nodePositions,
          canAddExistingConsensus: canAddExistingConsensus,
          graphs: graphs,
          selectedGraphId: selectedGraphId,
          canManageGraphs: canManageGraphs,
          onRefresh: onRefresh,
          onSelectGraph: onSelectGraph,
          onCreateGraph: onCreateGraph,
          onSelect: onSelect,
          onMoveNode: onMoveNode,
          onAddConsensus: onAddConsensus,
          onAddExistingConsensus: onAddExistingConsensus,
          onEditConsensus: onEditConsensus,
          onAddRelation: onAddRelation,
        );
        final detail = _DetailPanel(
          graph: graph,
          selectedId: selectedId,
          onSelect: onSelect,
          onEditConsensus: onEditConsensus,
          onRemoveNode: onRemoveNode,
          onRemoveEdge: onRemoveEdge,
        );
        if (isCompact) {
          return ListView(
            padding: const EdgeInsets.all(16),
            children: [
              SizedBox(height: 560, child: graphPanel),
              const SizedBox(height: 16),
              SizedBox(height: 520, child: detail),
            ],
          );
        }
        return Row(
          children: [
            Expanded(child: graphPanel),
            SizedBox(width: 360, child: detail),
          ],
        );
      },
    );
  }
}

class _GraphPanel extends StatelessWidget {
  const _GraphPanel({
    required this.graph,
    required this.selectedId,
    required this.isLoading,
    required this.error,
    required this.isSaving,
    required this.nodePositions,
    required this.canAddExistingConsensus,
    required this.graphs,
    required this.selectedGraphId,
    required this.canManageGraphs,
    required this.onRefresh,
    required this.onSelectGraph,
    required this.onCreateGraph,
    required this.onSelect,
    required this.onMoveNode,
    required this.onAddConsensus,
    required this.onAddExistingConsensus,
    required this.onEditConsensus,
    required this.onAddRelation,
  });

  final ConsensusGraph? graph;
  final String? selectedId;
  final bool isLoading;
  final Object? error;
  final bool isSaving;
  final Map<String, Offset> nodePositions;
  final bool canAddExistingConsensus;
  final List<ConsensusGraph> graphs;
  final String? selectedGraphId;
  final bool canManageGraphs;
  final VoidCallback onRefresh;
  final ValueChanged<String> onSelectGraph;
  final VoidCallback onCreateGraph;
  final ValueChanged<String> onSelect;
  final void Function(String id, Offset position) onMoveNode;
  final VoidCallback onAddConsensus;
  final VoidCallback onAddExistingConsensus;
  final VoidCallback onEditConsensus;
  final VoidCallback onAddRelation;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 12),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      graph?.name ?? '共识追溯图',
                      style: theme.textTheme.headlineSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      graph?.description ?? '记录团队共识的演进、依据和影响。',
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
              if (isSaving)
                const Padding(
                  padding: EdgeInsets.only(right: 8),
                  child: SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ),
              PopupMenuButton<String>(
                tooltip: '切换图谱',
                enabled: canManageGraphs && graphs.length > 1 && !isSaving,
                icon: const Icon(Icons.swap_horiz),
                onSelected: isSaving ? null : onSelectGraph,
                itemBuilder: (context) => [
                  for (final item in graphs)
                    PopupMenuItem(
                      value: item.id,
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              item.name,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                          if (item.id == selectedGraphId)
                            Icon(
                              Icons.check,
                              size: 18,
                              color: theme.colorScheme.primary,
                            ),
                        ],
                      ),
                    ),
                ],
              ),
              IconButton(
                tooltip: '新建图谱',
                onPressed: canManageGraphs && !isSaving ? onCreateGraph : null,
                icon: const Icon(Icons.addchart_outlined),
              ),
              IconButton(
                tooltip: '刷新',
                onPressed: isSaving ? null : onRefresh,
                icon: const Icon(Icons.refresh),
              ),
              IconButton(
                tooltip: '添加共识',
                onPressed: onAddConsensus,
                icon: const Icon(Icons.add_circle_outline),
              ),
              IconButton(
                tooltip: '纳入已有共识',
                onPressed: canAddExistingConsensus
                    ? onAddExistingConsensus
                    : null,
                icon: const Icon(Icons.playlist_add),
              ),
              IconButton(
                tooltip: '编辑共识',
                onPressed: onEditConsensus,
                icon: const Icon(Icons.edit_outlined),
              ),
              IconButton(
                tooltip: '添加关联',
                onPressed: onAddRelation,
                icon: const Icon(Icons.account_tree_outlined),
              ),
            ],
          ),
        ),
        if (error != null) _ErrorBanner(onRefresh: onRefresh),
        if (isLoading) const LinearProgressIndicator(minHeight: 2),
        Expanded(
          child: graph == null || graph!.nodes.isEmpty
              ? const _EmptyConsensusState()
              : _GraphCanvas(
                  graph: graph!,
                  selectedId: selectedId,
                  nodePositions: nodePositions,
                  onSelect: onSelect,
                  onMoveNode: onMoveNode,
                ),
        ),
      ],
    );
  }
}

class _GraphCanvas extends StatelessWidget {
  const _GraphCanvas({
    required this.graph,
    required this.selectedId,
    required this.nodePositions,
    required this.onSelect,
    required this.onMoveNode,
  });

  static const nodeSize = Size(196, 108);
  static const canvasSize = Size(1200, 760);

  final ConsensusGraph graph;
  final String? selectedId;
  final Map<String, Offset> nodePositions;
  final ValueChanged<String> onSelect;
  final void Function(String id, Offset position) onMoveNode;

  @override
  Widget build(BuildContext context) {
    final positions = _layoutNodes(graph);
    final theme = Theme.of(context);
    return DecoratedBox(
      decoration: BoxDecoration(
        color: theme.colorScheme.surfaceContainerLowest,
        border: Border(
          top: BorderSide(color: theme.colorScheme.outlineVariant),
        ),
      ),
      child: ClipRect(
        child: InteractiveViewer(
          minScale: 0.55,
          maxScale: 1.8,
          boundaryMargin: const EdgeInsets.all(120),
          constrained: false,
          child: SizedBox(
            width: canvasSize.width,
            height: canvasSize.height,
            child: Stack(
              children: [
                Positioned.fill(
                  child: CustomPaint(
                    painter: _GraphEdgesPainter(
                      graph: graph,
                      positions: positions,
                      selectedId: selectedId,
                    ),
                  ),
                ),
                for (final node in graph.nodes)
                  Positioned(
                    left: positions[node.id]!.dx,
                    top: positions[node.id]!.dy,
                    width: nodeSize.width,
                    height: nodeSize.height,
                    child: _DraggableGraphNode(
                      consensus: node,
                      isSelected: selectedId == node.id,
                      onTap: () => onSelect(node.id),
                      onMove: (delta) =>
                          onMoveNode(node.id, positions[node.id]! + delta),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Map<String, Offset> _layoutNodes(ConsensusGraph graph) {
    final result = <String, Offset>{};
    final levels = <String, int>{};
    final incoming = <String, List<String>>{};
    for (final node in graph.nodes) {
      incoming[node.id] = [];
    }
    for (final edge in graph.edges) {
      incoming[edge.to]?.add(edge.from);
    }
    for (final node in graph.nodes) {
      var level = 0;
      for (final parent in incoming[node.id] ?? const <String>[]) {
        level = math.max(level, (levels[parent] ?? 0) + 1);
      }
      levels[node.id] = level;
    }

    final byLevel = <int, List<Consensus>>{};
    for (final node in graph.nodes) {
      byLevel.putIfAbsent(levels[node.id] ?? 0, () => []).add(node);
    }
    for (final entry in byLevel.entries) {
      for (var index = 0; index < entry.value.length; index++) {
        final node = entry.value[index];
        result[node.id] =
            nodePositions[node.id] ??
            Offset(72 + entry.key * 270, 72 + index * 148);
      }
    }
    return result;
  }
}

class _GraphEdgesPainter extends CustomPainter {
  const _GraphEdgesPainter({
    required this.graph,
    required this.positions,
    required this.selectedId,
  });

  final ConsensusGraph graph;
  final Map<String, Offset> positions;
  final String? selectedId;

  @override
  void paint(Canvas canvas, Size size) {
    final line = Paint()
      ..strokeWidth = 2
      ..style = PaintingStyle.stroke;
    for (final edge in graph.edges) {
      final from = positions[edge.from];
      final to = positions[edge.to];
      if (from == null || to == null) {
        continue;
      }
      final start = Offset(
        from.dx + _GraphCanvas.nodeSize.width,
        from.dy + _GraphCanvas.nodeSize.height / 2,
      );
      final end = Offset(to.dx, to.dy + _GraphCanvas.nodeSize.height / 2);
      final highlighted = edge.from == selectedId || edge.to == selectedId;
      line.color = highlighted
          ? Colors.blue.shade700
          : Colors.blueGrey.shade400;
      final path = Path()
        ..moveTo(start.dx, start.dy)
        ..cubicTo(start.dx + 54, start.dy, end.dx - 54, end.dy, end.dx, end.dy);
      canvas.drawPath(path, line);

      final direction = end - start;
      final unit = direction / math.max(direction.distance, 1);
      final arrowBase = end - unit * 12;
      final perpendicular = Offset(-unit.dy, unit.dx) * 5;
      final arrow = Path()
        ..moveTo(end.dx, end.dy)
        ..lineTo(
          arrowBase.dx + perpendicular.dx,
          arrowBase.dy + perpendicular.dy,
        )
        ..lineTo(
          arrowBase.dx - perpendicular.dx,
          arrowBase.dy - perpendicular.dy,
        )
        ..close();
      canvas.drawPath(arrow, line..style = PaintingStyle.fill);

      final textPainter = TextPainter(
        text: TextSpan(
          text: edge.relationType,
          style: TextStyle(
            fontSize: 12,
            color: Colors.blueGrey.shade700,
            backgroundColor: Colors.white,
          ),
        ),
        textDirection: TextDirection.ltr,
      )..layout(maxWidth: 120);
      textPainter.paint(
        canvas,
        Offset(
          (start.dx + end.dx) / 2 - textPainter.width / 2,
          (start.dy + end.dy) / 2 - textPainter.height - 4,
        ),
      );
      line.style = PaintingStyle.stroke;
    }
  }

  @override
  bool shouldRepaint(covariant _GraphEdgesPainter oldDelegate) {
    return oldDelegate.graph != graph ||
        oldDelegate.selectedId != selectedId ||
        oldDelegate.positions != positions;
  }
}

class _DraggableGraphNode extends StatelessWidget {
  const _DraggableGraphNode({
    required this.consensus,
    required this.isSelected,
    required this.onTap,
    required this.onMove,
  });

  final Consensus consensus;
  final bool isSelected;
  final VoidCallback onTap;
  final ValueChanged<Offset> onMove;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final statusColor = _statusColor(theme, consensus.status);
    return GestureDetector(
      onTap: onTap,
      onPanUpdate: (details) => onMove(details.delta),
      child: Material(
        color: isSelected
            ? theme.colorScheme.primaryContainer
            : theme.colorScheme.surface,
        elevation: isSelected ? 5 : 2,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(14),
          decoration: BoxDecoration(
            border: Border.all(
              color: isSelected
                  ? theme.colorScheme.primary
                  : theme.colorScheme.outlineVariant,
              width: isSelected ? 2 : 1,
            ),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                width: 8,
                height: 64,
                decoration: BoxDecoration(
                  color: statusColor,
                  borderRadius: BorderRadius.circular(4),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      consensus.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.titleSmall?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      consensus.description.isEmpty
                          ? '暂无描述'
                          : consensus.description,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _DetailPanel extends StatelessWidget {
  const _DetailPanel({
    required this.graph,
    required this.selectedId,
    required this.onSelect,
    required this.onEditConsensus,
    required this.onRemoveNode,
    required this.onRemoveEdge,
  });

  final ConsensusGraph? graph;
  final String? selectedId;
  final ValueChanged<String> onSelect;
  final VoidCallback onEditConsensus;
  final VoidCallback onRemoveNode;
  final ValueChanged<ConsensusRelation> onRemoveEdge;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final selected = graph?.nodes
        .where((node) => node.id == selectedId)
        .firstOrNull;
    final incoming = selected == null
        ? const <ConsensusRelation>[]
        : graph!.edges.where((edge) => edge.to == selected.id).toList();
    final outgoing = selected == null
        ? const <ConsensusRelation>[]
        : graph!.edges.where((edge) => edge.from == selected.id).toList();

    return DecoratedBox(
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        border: Border(
          left: BorderSide(color: theme.colorScheme.outlineVariant),
        ),
      ),
      child: ListView(
        padding: const EdgeInsets.all(20),
        children: [
          Text('图谱概览', style: theme.textTheme.titleMedium),
          const SizedBox(height: 12),
          _DetailRow(label: '节点', value: '${graph?.nodes.length ?? 0}'),
          _DetailRow(label: '关联', value: '${graph?.edges.length ?? 0}'),
          const Divider(height: 28),
          if (selected == null)
            Text(
              '选择一个节点查看详情，或使用工具栏创建新的共识和关联。',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            )
          else ...[
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: Text(
                    selected.title,
                    style: theme.textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                IconButton(
                  tooltip: '编辑共识',
                  onPressed: selected.status == 'proposed'
                      ? onEditConsensus
                      : null,
                  icon: const Icon(Icons.edit_outlined),
                ),
              ],
            ),
            const SizedBox(height: 10),
            Text(selected.description.isEmpty ? '暂无描述' : selected.description),
            const SizedBox(height: 18),
            _StatusPill(status: selected.status),
            const SizedBox(height: 18),
            _DetailRow(label: '创建时间', value: selected.createdAt),
            _DetailRow(label: '更新时间', value: selected.updatedAt),
            const SizedBox(height: 10),
            OutlinedButton.icon(
              onPressed: onRemoveNode,
              icon: const Icon(Icons.remove_circle_outline),
              label: const Text('移出当前图'),
            ),
            if (incoming.isNotEmpty) ...[
              const SizedBox(height: 24),
              Text('前置关联', style: theme.textTheme.titleSmall),
              const SizedBox(height: 8),
              for (final edge in incoming)
                _RelationRow(
                  relation: edge,
                  counterpart: _nodeTitle(graph!, edge.from),
                  onSelect: () => onSelect(edge.from),
                  onRemove: () => onRemoveEdge(edge),
                ),
            ],
            if (outgoing.isNotEmpty) ...[
              const SizedBox(height: 24),
              Text('后续关联', style: theme.textTheme.titleSmall),
              const SizedBox(height: 8),
              for (final edge in outgoing)
                _RelationRow(
                  relation: edge,
                  counterpart: _nodeTitle(graph!, edge.to),
                  onSelect: () => onSelect(edge.to),
                  onRemove: () => onRemoveEdge(edge),
                ),
            ],
          ],
        ],
      ),
    );
  }
}

class _RelationRow extends StatelessWidget {
  const _RelationRow({
    required this.relation,
    required this.counterpart,
    required this.onSelect,
    required this.onRemove,
  });

  final ConsensusRelation relation;
  final String counterpart;
  final VoidCallback onSelect;
  final VoidCallback onRemove;

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: EdgeInsets.zero,
      dense: true,
      title: Text(counterpart, maxLines: 2, overflow: TextOverflow.ellipsis),
      subtitle: Text(relation.relationType),
      onTap: onSelect,
      trailing: IconButton(
        tooltip: '移除关联',
        onPressed: onRemove,
        icon: const Icon(Icons.close),
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  const _DetailRow({required this.label, required this.value});

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 72,
            child: Text(
              label,
              style: theme.textTheme.bodySmall?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          Expanded(child: Text(value.isEmpty ? '-' : value)),
        ],
      ),
    );
  }
}

class _StatusPill extends StatelessWidget {
  const _StatusPill({required this.status});

  final String status;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = _statusColor(theme, status);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        border: Border.all(color: color),
        borderRadius: BorderRadius.circular(999),
      ),
      child: Text(
        _statusLabel(status),
        style: theme.textTheme.labelSmall?.copyWith(color: color),
      ),
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({required this.onRefresh});

  final VoidCallback onRefresh;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.fromLTRB(20, 0, 20, 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: theme.colorScheme.errorContainer,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          Icon(Icons.cloud_off, color: theme.colorScheme.onErrorContainer),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              'Provider 未连接',
              style: TextStyle(color: theme.colorScheme.onErrorContainer),
            ),
          ),
          TextButton(onPressed: onRefresh, child: const Text('重试')),
        ],
      ),
    );
  }
}

class _EmptyConsensusState extends StatelessWidget {
  const _EmptyConsensusState();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Center(
      child: Text(
        '暂无共识节点，点击“添加共识”开始记录。',
        style: theme.textTheme.titleMedium?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
      ),
    );
  }
}

class _ConsensusDraft {
  const _ConsensusDraft({required this.title, required this.description});

  final String title;
  final String description;
}

class _GraphDraft {
  const _GraphDraft({required this.name, required this.description});

  final String name;
  final String description;
}

class _RelationDraft {
  const _RelationDraft({
    required this.from,
    required this.to,
    required this.relationType,
  });

  final String from;
  final String to;
  final String relationType;
}

Future<_GraphDraft?> _showGraphDialog(BuildContext context) async {
  final nameController = TextEditingController();
  final descriptionController = TextEditingController();
  final formKey = GlobalKey<FormState>();
  final result = await showDialog<_GraphDraft>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('新建图谱'),
      content: Form(
        key: formKey,
        child: SizedBox(
          width: 440,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: nameController,
                autofocus: true,
                decoration: const InputDecoration(
                  labelText: '图谱名称',
                  border: OutlineInputBorder(),
                ),
                validator: (value) =>
                    value == null || value.trim().isEmpty ? '请输入图谱名称' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: descriptionController,
                minLines: 3,
                maxLines: 6,
                decoration: const InputDecoration(
                  labelText: '说明',
                  border: OutlineInputBorder(),
                ),
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: () {
            if (!formKey.currentState!.validate()) {
              return;
            }
            Navigator.pop(
              context,
              _GraphDraft(
                name: nameController.text.trim(),
                description: descriptionController.text.trim(),
              ),
            );
          },
          child: const Text('创建'),
        ),
      ],
    ),
  );
  nameController.dispose();
  descriptionController.dispose();
  return result;
}

Future<_ConsensusDraft?> _showConsensusDialog(
  BuildContext context, {
  Consensus? initial,
}) async {
  final titleController = TextEditingController(text: initial?.title);
  final descriptionController = TextEditingController(
    text: initial?.description,
  );
  final formKey = GlobalKey<FormState>();
  final result = await showDialog<_ConsensusDraft>(
    context: context,
    builder: (context) => AlertDialog(
      title: Text(initial == null ? '添加共识' : '编辑共识'),
      content: Form(
        key: formKey,
        child: SizedBox(
          width: 440,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: titleController,
                autofocus: true,
                decoration: const InputDecoration(
                  labelText: '标题',
                  border: OutlineInputBorder(),
                ),
                validator: (value) =>
                    value == null || value.trim().isEmpty ? '请输入标题' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: descriptionController,
                minLines: 3,
                maxLines: 6,
                decoration: const InputDecoration(
                  labelText: '描述',
                  border: OutlineInputBorder(),
                ),
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: () {
            if (!formKey.currentState!.validate()) {
              return;
            }
            Navigator.pop(
              context,
              _ConsensusDraft(
                title: titleController.text.trim(),
                description: descriptionController.text.trim(),
              ),
            );
          },
          child: const Text('保存'),
        ),
      ],
    ),
  );
  titleController.dispose();
  descriptionController.dispose();
  return result;
}

Future<_RelationDraft?> _showRelationDialog(
  BuildContext context,
  List<Consensus> nodes,
) async {
  var from = nodes.first.id;
  var to = nodes[1].id;
  final typeController = TextEditingController(text: '支持');
  final formKey = GlobalKey<FormState>();
  final result = await showDialog<_RelationDraft>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('添加关联'),
      content: Form(
        key: formKey,
        child: SizedBox(
          width: 440,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField<String>(
                initialValue: from,
                decoration: const InputDecoration(
                  labelText: '起点',
                  border: OutlineInputBorder(),
                ),
                items: [
                  for (final node in nodes)
                    DropdownMenuItem(value: node.id, child: Text(node.title)),
                ],
                onChanged: (value) => from = value ?? from,
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<String>(
                initialValue: to,
                decoration: const InputDecoration(
                  labelText: '终点',
                  border: OutlineInputBorder(),
                ),
                items: [
                  for (final node in nodes)
                    DropdownMenuItem(value: node.id, child: Text(node.title)),
                ],
                onChanged: (value) => to = value ?? to,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: typeController,
                decoration: const InputDecoration(
                  labelText: '关联类型',
                  helperText: '可填写前置条件、支持、反对、补充或自定义语义。',
                  border: OutlineInputBorder(),
                ),
                validator: (value) =>
                    value == null || value.trim().isEmpty ? '请输入关联类型' : null,
              ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: () {
            if (!formKey.currentState!.validate() || from == to) {
              return;
            }
            Navigator.pop(
              context,
              _RelationDraft(
                from: from,
                to: to,
                relationType: typeController.text.trim(),
              ),
            );
          },
          child: const Text('建立关联'),
        ),
      ],
    ),
  );
  typeController.dispose();
  return result;
}

Future<String?> _showExistingConsensusDialog(
  BuildContext context,
  List<Consensus> consensuses,
) {
  var selectedId = consensuses.first.id;
  return showDialog<String>(
    context: context,
    builder: (context) => AlertDialog(
      title: const Text('纳入已有共识'),
      content: SizedBox(
        width: 440,
        child: DropdownButtonFormField<String>(
          initialValue: selectedId,
          decoration: const InputDecoration(
            labelText: '共识',
            border: OutlineInputBorder(),
          ),
          items: [
            for (final consensus in consensuses)
              DropdownMenuItem(
                value: consensus.id,
                child: Text(consensus.title, overflow: TextOverflow.ellipsis),
              ),
          ],
          onChanged: (value) => selectedId = value ?? selectedId,
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        FilledButton(
          onPressed: () => Navigator.pop(context, selectedId),
          child: const Text('纳入'),
        ),
      ],
    ),
  );
}

String _nodeTitle(ConsensusGraph graph, String id) {
  return graph.nodes
          .where((node) => node.id == id)
          .map((node) => node.title)
          .firstOrNull ??
      id;
}

Color _statusColor(ThemeData theme, String status) {
  return switch (status) {
    'confirmed' => Colors.green.shade700,
    'deprecated' => theme.colorScheme.error,
    _ => theme.colorScheme.primary,
  };
}

String _statusLabel(String status) {
  return switch (status) {
    'confirmed' => '已确认',
    'deprecated' => '已废弃',
    _ => '提议中',
  };
}
