import 'package:flutter/material.dart';

import '../models/consensus.dart';
import '../services/consensus_api.dart';

typedef ConsensusLoader = Future<List<Consensus>> Function();

class ConsensusTraceabilityScreen extends StatefulWidget {
  const ConsensusTraceabilityScreen({
    super.key,
    this.loadConsensuses = _loadConsensuses,
  });

  final ConsensusLoader loadConsensuses;

  static Future<List<Consensus>> _loadConsensuses() {
    return const ConsensusApiClient().listConsensuses();
  }

  @override
  State<ConsensusTraceabilityScreen> createState() =>
      _ConsensusTraceabilityScreenState();
}

class _ConsensusTraceabilityScreenState
    extends State<ConsensusTraceabilityScreen> {
  late Future<List<Consensus>> _consensuses;
  Consensus? _selected;

  @override
  void initState() {
    super.initState();
    _consensuses = widget.loadConsensuses();
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
              child: FutureBuilder<List<Consensus>>(
                future: _consensuses,
                builder: (context, snapshot) {
                  final items = snapshot.data ?? const <Consensus>[];
                  final selected = _selected;

                  return _ConsensusSurface(
                    items: items,
                    selected: selected,
                    isLoading:
                        snapshot.connectionState == ConnectionState.waiting,
                    hasError: snapshot.hasError,
                    onRefresh: _refresh,
                    onSelect: (item) => setState(() => _selected = item),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _refresh() {
    setState(() {
      _selected = null;
      _consensuses = widget.loadConsensuses();
    });
  }
}

class _ConsensusSurface extends StatelessWidget {
  const _ConsensusSurface({
    required this.items,
    required this.selected,
    required this.isLoading,
    required this.hasError,
    required this.onRefresh,
    required this.onSelect,
  });

  final List<Consensus> items;
  final Consensus? selected;
  final bool isLoading;
  final bool hasError;
  final VoidCallback onRefresh;
  final ValueChanged<Consensus> onSelect;

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        final isCompact = constraints.maxWidth < 820;
        final graph = _GraphPanel(
          items: items,
          selected: selected,
          isLoading: isLoading,
          hasError: hasError,
          onRefresh: onRefresh,
          onSelect: onSelect,
        );
        final detail = _DetailPanel(consensus: selected, total: items.length);

        if (isCompact) {
          return ListView(
            padding: const EdgeInsets.all(20),
            children: [
              SizedBox(height: 520, child: graph),
              const SizedBox(height: 16),
              detail,
            ],
          );
        }

        return Row(
          children: [
            Expanded(child: graph),
            SizedBox(width: 360, child: detail),
          ],
        );
      },
    );
  }
}

class _GraphPanel extends StatelessWidget {
  const _GraphPanel({
    required this.items,
    required this.selected,
    required this.isLoading,
    required this.hasError,
    required this.onRefresh,
    required this.onSelect,
  });

  final List<Consensus> items;
  final Consensus? selected;
  final bool isLoading;
  final bool hasError;
  final VoidCallback onRefresh;
  final ValueChanged<Consensus> onSelect;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('共识追溯图', style: theme.textTheme.headlineMedium),
                    const SizedBox(height: 6),
                    Text(
                      'Message -> Consensus -> Memo',
                      style: theme.textTheme.bodyMedium?.copyWith(
                        color: theme.colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
              IconButton(
                tooltip: '刷新',
                onPressed: onRefresh,
                icon: const Icon(Icons.refresh),
              ),
            ],
          ),
          const SizedBox(height: 20),
          if (hasError) _ErrorBanner(onRefresh: onRefresh),
          if (isLoading) const LinearProgressIndicator(minHeight: 2),
          const SizedBox(height: 16),
          Expanded(
            child: items.isEmpty && !isLoading
                ? const _EmptyConsensusState()
                : ListView.separated(
                    itemCount: items.length,
                    separatorBuilder: (context, index) =>
                        const _TraceConnector(),
                    itemBuilder: (context, index) {
                      final item = items[index];
                      return _ConsensusNode(
                        consensus: item,
                        isSelected: selected?.id == item.id,
                        onTap: () => onSelect(item),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}

class _ConsensusNode extends StatelessWidget {
  const _ConsensusNode({
    required this.consensus,
    required this.isSelected,
    required this.onTap,
  });

  final Consensus consensus;
  final bool isSelected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = _statusColor(theme, consensus.status);

    return Material(
      color: isSelected
          ? theme.colorScheme.primaryContainer
          : theme.colorScheme.surfaceContainerHighest,
      borderRadius: BorderRadius.circular(8),
      child: InkWell(
        borderRadius: BorderRadius.circular(8),
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Container(
                width: 10,
                height: 54,
                decoration: BoxDecoration(
                  color: color,
                  borderRadius: BorderRadius.circular(999),
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      consensus.title,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: theme.textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w700,
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      consensus.description.isEmpty
                          ? '暂无描述'
                          : consensus.description,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              _StatusPill(status: consensus.status),
            ],
          ),
        ),
      ),
    );
  }
}

class _TraceConnector extends StatelessWidget {
  const _TraceConnector();

  @override
  Widget build(BuildContext context) {
    return const Padding(
      padding: EdgeInsets.only(left: 21),
      child: SizedBox(height: 22, child: VerticalDivider(thickness: 2)),
    );
  }
}

class _DetailPanel extends StatelessWidget {
  const _DetailPanel({required this.consensus, required this.total});

  final Consensus? consensus;
  final int total;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return DecoratedBox(
      decoration: BoxDecoration(
        color: theme.colorScheme.surface,
        border: Border(
          left: BorderSide(color: theme.colorScheme.outlineVariant),
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('共识总数', style: theme.textTheme.labelLarge),
            const SizedBox(height: 8),
            Text('$total', style: theme.textTheme.displaySmall),
            const SizedBox(height: 24),
            Text('当前共识', style: theme.textTheme.titleMedium),
            const SizedBox(height: 12),
            if (consensus == null)
              Text(
                '暂无共识记录',
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              )
            else ...[
              Text(
                consensus!.title,
                style: theme.textTheme.titleLarge?.copyWith(
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 12),
              Text(
                consensus!.description.isEmpty
                    ? '暂无描述'
                    : consensus!.description,
              ),
              const SizedBox(height: 20),
              _DetailRow(label: '状态', value: _statusLabel(consensus!.status)),
              _DetailRow(label: '创建时间', value: consensus!.createdAt),
              _DetailRow(label: '更新时间', value: consensus!.updatedAt),
            ],
          ],
        ),
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
      margin: const EdgeInsets.only(bottom: 8),
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
        '暂无共识记录',
        style: theme.textTheme.titleMedium?.copyWith(
          color: theme.colorScheme.onSurfaceVariant,
        ),
      ),
    );
  }
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
