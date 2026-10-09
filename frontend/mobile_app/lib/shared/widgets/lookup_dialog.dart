import 'package:flutter/material.dart';

/// Displays searchable lookup options and returns the selected item(s).
class LookupDialog<T> extends StatefulWidget {
  const LookupDialog({
    super.key,
    required this.items,
    required this.title,
    required this.labelBuilder,
    this.subtitleBuilder,
    this.selectedItems = const [],
    this.multiSelect = false,
    this.searchHint = 'Cari...',
    this.emptyMessage = 'Data tidak ditemukan',
    this.isLoading = false,
    this.errorMessage,
    this.maxWidth = 560,
  });

  final List<T> items;
  final String title;
  final String Function(T item) labelBuilder;
  final String? Function(T item)? subtitleBuilder;
  final List<T> selectedItems;
  final bool multiSelect;
  final String searchHint;
  final String emptyMessage;
  final bool isLoading;
  final String? errorMessage;
  final double maxWidth;

  /// Opens this lookup and returns one item, selected items, or null if closed.
  static Future<Object?> show<T>(
    BuildContext context, {
    required List<T> items,
    required String title,
    required String Function(T item) labelBuilder,
    String? Function(T item)? subtitleBuilder,
    List<T> selectedItems = const [],
    bool multiSelect = false,
    String searchHint = 'Cari...',
    String emptyMessage = 'Data tidak ditemukan',
    bool isLoading = false,
    String? errorMessage,
  }) {
    return showDialog<Object?>(
      context: context,
      builder: (_) => LookupDialog<T>(
        items: items,
        title: title,
        labelBuilder: labelBuilder,
        subtitleBuilder: subtitleBuilder,
        selectedItems: selectedItems,
        multiSelect: multiSelect,
        searchHint: searchHint,
        emptyMessage: emptyMessage,
        isLoading: isLoading,
        errorMessage: errorMessage,
      ),
    );
  }

  @override
  State<LookupDialog<T>> createState() => _LookupDialogState<T>();
}

class _LookupDialogState<T> extends State<LookupDialog<T>> {
  final _searchController = TextEditingController();
  late final Set<T> _selected = {...widget.selectedItems};

  List<T> get _filteredItems {
    final search = _searchController.text.trim().toLowerCase();
    if (search.isEmpty) return widget.items;
    return widget.items
        .where(
          (item) =>
              widget.labelBuilder(item).toLowerCase().contains(search) ||
              (widget.subtitleBuilder
                      ?.call(item)
                      ?.toLowerCase()
                      .contains(search) ??
                  false),
        )
        .toList(growable: false);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _select(T item) {
    if (widget.multiSelect) {
      setState(() {
        if (!_selected.add(item)) _selected.remove(item);
      });
      return;
    }
    Navigator.of(context).pop<T>(item);
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filteredItems;
    return AlertDialog(
      title: Text(widget.title),
      content: SizedBox(
        width: widget.maxWidth,
        height: 420,
        child: Column(
          children: [
            TextField(
              controller: _searchController,
              onChanged: (_) => setState(() {}),
              decoration: InputDecoration(
                hintText: widget.searchHint,
                prefixIcon: const Icon(Icons.search),
                border: const OutlineInputBorder(),
              ),
            ),
            const SizedBox(height: 12),
            Expanded(child: _buildResults(filtered)),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Batal'),
        ),
        if (widget.multiSelect)
          FilledButton(
            onPressed: () =>
                Navigator.of(context).pop<List<T>>(_selected.toList()),
            child: Text('Pilih (${_selected.length})'),
          ),
      ],
    );
  }

  Widget _buildResults(List<T> filtered) {
    if (widget.isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (widget.errorMessage != null) {
      return Center(
        child: Text(widget.errorMessage!, textAlign: TextAlign.center),
      );
    }
    if (filtered.isEmpty) {
      return Center(child: Text(widget.emptyMessage));
    }

    return ListView.builder(
      itemCount: filtered.length,
      itemBuilder: (context, index) {
        final item = filtered[index];
        final selected = _selected.contains(item);
        final subtitle = widget.subtitleBuilder?.call(item);
        return ListTile(
          title: Text(widget.labelBuilder(item)),
          subtitle: subtitle == null ? null : Text(subtitle),
          leading: widget.multiSelect
              ? Checkbox(value: selected, onChanged: (_) => _select(item))
              : null,
          trailing: widget.multiSelect || !selected
              ? null
              : const Icon(Icons.check),
          onTap: () => _select(item),
        );
      },
    );
  }
}
