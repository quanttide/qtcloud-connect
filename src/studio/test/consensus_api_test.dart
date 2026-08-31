import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:qtcloud_connect_studio/services/consensus_api.dart';

void main() {
  test('listConsensuses fetches every provider page', () async {
    final requestedPages = <int>[];
    final client = MockClient((request) async {
      final page = int.parse(request.url.queryParameters['page']!);
      final pageSize = int.parse(request.url.queryParameters['page_size']!);
      requestedPages.add(page);

      final items = page == 1
          ? List.generate(
              pageSize,
              (index) => _consensusJson('consensus-$index'),
            )
          : [_consensusJson('consensus-$pageSize')];

      return http.Response(
        jsonEncode({
          'items': items,
          'total': pageSize + 1,
          'page': page,
          'page_size': pageSize,
        }),
        200,
        headers: const {'content-type': 'application/json'},
      );
    });

    final api = ConsensusApiClient(
      endpoint: 'http://provider.test/api',
      client: client,
    );

    final consensuses = await api.listConsensuses();

    expect(consensuses, hasLength(101));
    expect(requestedPages, [1, 2]);
  });
}

Map<String, dynamic> _consensusJson(String id) {
  return {
    'id': id,
    'title': '共识 $id',
    'description': '描述 $id',
    'status': 'proposed',
    'created_at': '2026-08-30T00:00:00Z',
    'updated_at': '2026-08-30T00:00:00Z',
  };
}
