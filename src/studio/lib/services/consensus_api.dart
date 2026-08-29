import 'dart:convert';

import 'package:http/http.dart' as http;

import '../models/consensus.dart';

class ConsensusApiClient {
  const ConsensusApiClient({this.endpoint = defaultEndpoint});

  static const defaultEndpoint = String.fromEnvironment(
    'CONNECT_PROVIDER_ENDPOINT',
    defaultValue: 'http://localhost:8000/api',
  );
  final String endpoint;

  Future<List<Consensus>> listConsensuses() async {
    final uri = Uri.parse(
      '${endpoint.replaceAll(RegExp(r'/+$'), '')}/consensuses',
    );
    final response = await http.get(uri);
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw ConsensusApiException('Provider returned ${response.statusCode}');
    }

    final decoded = jsonDecode(response.body);
    final items = switch (decoded) {
      {'items': final List<dynamic> values} => values,
      final List<dynamic> values => values,
      _ => throw const FormatException('Unexpected consensus response'),
    };

    return items
        .whereType<Map<String, dynamic>>()
        .map(Consensus.fromJson)
        .where((item) => item.id.isNotEmpty && item.title.isNotEmpty)
        .toList(growable: false);
  }

  Future<List<ConsensusGraph>> listGraphs() async {
    final decoded = await _request('GET', '/consensus-graphs');
    final items = switch (decoded) {
      {'items': final List<dynamic> values} => values,
      final List<dynamic> values => values,
      _ => throw const FormatException('Unexpected graph response'),
    };
    return items
        .whereType<Map<String, dynamic>>()
        .map(ConsensusGraph.fromJson)
        .where((item) => item.id.isNotEmpty)
        .toList(growable: false);
  }

  Future<ConsensusGraph> getGraph(String id) async {
    final decoded = await _request('GET', '/consensus-graphs/$id');
    return ConsensusGraph.fromJson(decoded);
  }

  Future<ConsensusGraph> createGraph({
    required String name,
    String description = '',
  }) async {
    final decoded = await _request(
      'POST',
      '/consensus-graphs',
      body: {'name': name, 'description': description},
    );
    return ConsensusGraph.fromJson(decoded);
  }

  Future<Consensus> createConsensus({
    required String title,
    required String description,
  }) async {
    final decoded = await _request(
      'POST',
      '/consensuses',
      body: {'title': title, 'description': description},
    );
    return Consensus.fromJson(decoded);
  }

  Future<Consensus> updateConsensus({
    required String id,
    required String title,
    required String description,
  }) async {
    final decoded = await _request(
      'PUT',
      '/consensuses/$id',
      body: {'title': title, 'description': description},
    );
    return Consensus.fromJson(decoded);
  }

  Future<ConsensusGraph> addNode({
    required String graphId,
    required String consensusId,
  }) async {
    final decoded = await _request(
      'POST',
      '/consensus-graphs/$graphId/nodes',
      body: {'consensus_id': consensusId},
    );
    return ConsensusGraph.fromJson(decoded);
  }

  Future<ConsensusGraph> removeNode({
    required String graphId,
    required String consensusId,
  }) async {
    final decoded = await _request(
      'DELETE',
      '/consensus-graphs/$graphId/nodes/$consensusId',
    );
    return ConsensusGraph.fromJson(decoded);
  }

  Future<ConsensusRelation> createRelation({
    required String from,
    required String to,
    required String relationType,
  }) async {
    final decoded = await _request(
      'POST',
      '/consensus-relations',
      body: {'from': from, 'to': to, 'relation_type': relationType},
    );
    return ConsensusRelation.fromJson(decoded);
  }

  Future<ConsensusGraph> addEdge({
    required String graphId,
    required String relationId,
  }) async {
    final decoded = await _request(
      'POST',
      '/consensus-graphs/$graphId/edges',
      body: {'relation_id': relationId},
    );
    return ConsensusGraph.fromJson(decoded);
  }

  Future<ConsensusGraph> createGraphRelation({
    required String graphId,
    required String from,
    required String to,
    required String relationType,
  }) async {
    final decoded = await _request(
      'POST',
      '/consensus-graphs/$graphId/relations',
      body: {'from': from, 'to': to, 'relation_type': relationType},
    );
    return ConsensusGraph.fromJson(decoded);
  }

  Future<ConsensusGraph> removeEdge({
    required String graphId,
    required String relationId,
  }) async {
    final decoded = await _request(
      'DELETE',
      '/consensus-graphs/$graphId/edges/$relationId',
    );
    return ConsensusGraph.fromJson(decoded);
  }

  Future<Map<String, dynamic>> _request(
    String method,
    String path, {
    Map<String, Object?>? body,
  }) async {
    final uri = Uri.parse(
      '${endpoint.replaceAll(RegExp(r'/+$'), '')}/'
      '${path.replaceAll(RegExp(r'^/+'), '')}',
    );
    final encodedBody = body == null ? null : jsonEncode(body);
    final response = switch (method) {
      'GET' => await http.get(uri),
      'POST' => await http.post(
        uri,
        headers: const {'Content-Type': 'application/json'},
        body: encodedBody,
      ),
      'PUT' => await http.put(
        uri,
        headers: const {'Content-Type': 'application/json'},
        body: encodedBody,
      ),
      'DELETE' => await http.delete(uri),
      _ => throw ArgumentError.value(
        method,
        'method',
        'Unsupported HTTP method',
      ),
    };
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw ConsensusApiException(
        'Provider returned ${response.statusCode}: ${response.body}',
      );
    }
    if (response.body.trim().isEmpty) {
      return const <String, dynamic>{};
    }
    final decoded = jsonDecode(response.body);
    if (decoded is! Map<String, dynamic>) {
      throw const FormatException('Unexpected object response');
    }
    return decoded;
  }
}

class ConsensusApiException implements Exception {
  const ConsensusApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
