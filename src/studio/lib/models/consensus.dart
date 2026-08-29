class Consensus {
  const Consensus({
    required this.id,
    required this.title,
    required this.description,
    required this.createdAt,
    required this.updatedAt,
    this.status = 'proposed',
  });

  final String id;
  final String title;
  final String description;
  final String status;
  final String createdAt;
  final String updatedAt;

  factory Consensus.fromJson(Map<String, dynamic> json) {
    return Consensus(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      status: json['status'] as String? ?? 'proposed',
      createdAt: json['created_at'] as String? ?? '',
      updatedAt: json['updated_at'] as String? ?? '',
    );
  }

  Consensus copyWith({
    String? title,
    String? description,
    String? status,
    String? updatedAt,
  }) {
    return Consensus(
      id: id,
      title: title ?? this.title,
      description: description ?? this.description,
      status: status ?? this.status,
      createdAt: createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}

class ConsensusRelation {
  const ConsensusRelation({
    required this.id,
    required this.from,
    required this.to,
    required this.relationType,
  });

  final String id;
  final String from;
  final String to;
  final String relationType;

  factory ConsensusRelation.fromJson(Map<String, dynamic> json) {
    return ConsensusRelation(
      id: json['id'] as String? ?? '',
      from: json['from'] as String? ?? '',
      to: json['to'] as String? ?? '',
      relationType: json['relation_type'] as String? ?? '',
    );
  }
}

class ConsensusGraph {
  const ConsensusGraph({
    required this.id,
    required this.name,
    required this.description,
    required this.nodes,
    required this.edges,
    required this.createdAt,
    required this.updatedAt,
  });

  final String id;
  final String name;
  final String description;
  final List<Consensus> nodes;
  final List<ConsensusRelation> edges;
  final String createdAt;
  final String updatedAt;

  factory ConsensusGraph.fromJson(Map<String, dynamic> json) {
    return ConsensusGraph(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      nodes: _decodeList(json['nodes'], Consensus.fromJson),
      edges: _decodeList(json['edges'], ConsensusRelation.fromJson),
      createdAt: json['created_at'] as String? ?? '',
      updatedAt: json['updated_at'] as String? ?? '',
    );
  }

  ConsensusGraph copyWith({
    String? name,
    String? description,
    List<Consensus>? nodes,
    List<ConsensusRelation>? edges,
    String? updatedAt,
  }) {
    return ConsensusGraph(
      id: id,
      name: name ?? this.name,
      description: description ?? this.description,
      nodes: nodes ?? this.nodes,
      edges: edges ?? this.edges,
      createdAt: createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}

List<T> _decodeList<T>(
  Object? value,
  T Function(Map<String, dynamic>) decode,
) {
  if (value is! List<dynamic>) {
    return const <Never>[];
  }
  return value
      .whereType<Map<String, dynamic>>()
      .map(decode)
      .toList(growable: false);
}
