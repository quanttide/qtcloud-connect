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
}
