import 'package:flutter_test/flutter_test.dart';
import 'package:qtcloud_connect_studio/services/consensus_api.dart';

void main() {
  test('Studio client reads data from the v0.1 Provider', () async {
    const endpoint = String.fromEnvironment('CONNECT_PROVIDER_ENDPOINT');
    if (endpoint.isEmpty) {
      return;
    }

    final api = ConsensusApiClient(endpoint: endpoint);
    final graphs = await api.listGraphs();
    expect(graphs, isNotEmpty);

    final graph = await api.getGraph(graphs.first.id);
    expect(graph.nodes.map((node) => node.title), contains('v0.1 验收共识（更新）'));
    expect(graph.nodes.map((node) => node.title), contains('v0.1 第二条验收共识'));
    expect(graph.edges, hasLength(1));
  });
}
