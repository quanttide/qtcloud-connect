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
}

class ConsensusApiException implements Exception {
  const ConsensusApiException(this.message);

  final String message;

  @override
  String toString() => message;
}
