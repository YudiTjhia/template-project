import 'package:flutter/material.dart';

enum EditFieldType {
  text,
  email,
  password,
  number,
  multiline,
  dropdown,
  multiSelect,
}

class EditFieldOption {
  const EditFieldOption({required this.value, required this.label});

  final String value;
  final String label;
}

/// Metadata used by [EditFields] to render and validate one form field.
class EditFieldMetadata {
  const EditFieldMetadata({
    required this.name,
    required this.label,
    this.type = EditFieldType.text,
    this.initialValue,
    this.placeholder,
    this.helperText,
    this.required = false,
    this.enabled = true,
    this.readOnly = false,
    this.obscureText = false,
    this.maxLines,
    this.options = const [],
    this.initialValues = const [],
    this.controller,
    this.validator,
    this.onChanged,
    this.onMultiChanged,
  });

  final String name;
  final String label;
  final EditFieldType type;
  final String? initialValue;
  final String? placeholder;
  final String? helperText;
  final bool required;
  final bool enabled;
  final bool readOnly;
  final bool obscureText;
  final int? maxLines;
  final List<EditFieldOption> options;
  final List<String> initialValues;
  final TextEditingController? controller;
  final String? Function(String?)? validator;
  final ValueChanged<String?>? onChanged;
  final ValueChanged<List<String>>? onMultiChanged;

  EditFieldMetadata copyWith({bool? enabled, bool? readOnly}) {
    return EditFieldMetadata(
      name: name,
      label: label,
      type: type,
      initialValue: initialValue,
      placeholder: placeholder,
      helperText: helperText,
      required: required,
      enabled: enabled ?? this.enabled,
      readOnly: readOnly ?? this.readOnly,
      obscureText: obscureText,
      maxLines: maxLines,
      options: options,
      initialValues: initialValues,
      controller: controller,
      validator: validator,
      onChanged: onChanged,
      onMultiChanged: onMultiChanged,
    );
  }
}

/// Renders form fields from metadata. Place this widget inside a [Form].
class EditFields extends StatelessWidget {
  const EditFields({super.key, required this.fields, this.spacing = 16});

  final List<EditFieldMetadata> fields;
  final double spacing;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        for (var index = 0; index < fields.length; index++) ...[
          if (index > 0) SizedBox(height: spacing),
          _buildField(fields[index]),
        ],
      ],
    );
  }

  Widget _buildField(EditFieldMetadata field) {
    String? validate(String? value) {
      if (field.required && (value == null || value.trim().isEmpty)) {
        return '${field.label} wajib diisi';
      }
      return field.validator?.call(value);
    }

    final decoration = InputDecoration(
      labelText: field.label,
      hintText: field.placeholder,
      helperText: field.helperText,
    );

    if (field.type == EditFieldType.dropdown) {
      return DropdownButtonFormField<String>(
        key: ValueKey(field.name),
        initialValue: field.initialValue,
        decoration: decoration,
        items: [
          for (final option in field.options)
            DropdownMenuItem(value: option.value, child: Text(option.label)),
        ],
        validator: validate,
        onChanged: field.enabled && !field.readOnly ? field.onChanged : null,
      );
    }

    if (field.type == EditFieldType.multiSelect) {
      return FormField<List<String>>(
        key: ValueKey(field.name),
        initialValue: field.initialValues,
        validator: (values) {
          if (field.required && (values == null || values.isEmpty)) {
            return '${field.label} wajib dipilih';
          }
          return field.validator?.call(values?.join(','));
        },
        builder: (state) => InputDecorator(
          decoration: decoration.copyWith(errorText: state.errorText),
          child: Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              for (final option in field.options)
                FilterChip(
                  label: Text(option.label),
                  selected: state.value?.contains(option.value) ?? false,
                  onSelected: !field.enabled || field.readOnly
                      ? null
                      : (selected) {
                          final values = [...?state.value];
                          if (selected) {
                            values.add(option.value);
                          } else {
                            values.remove(option.value);
                          }
                          state.didChange(values);
                          field.onMultiChanged?.call(values);
                        },
                ),
            ],
          ),
        ),
      );
    }

    return TextFormField(
      key: ValueKey(field.name),
      controller: field.controller,
      initialValue: field.controller == null ? field.initialValue : null,
      enabled: field.enabled,
      readOnly: field.readOnly,
      obscureText: field.obscureText || field.type == EditFieldType.password,
      maxLines: field.type == EditFieldType.password
          ? 1
          : field.maxLines ?? (field.type == EditFieldType.multiline ? 4 : 1),
      keyboardType: switch (field.type) {
        EditFieldType.email => TextInputType.emailAddress,
        EditFieldType.number => TextInputType.number,
        EditFieldType.multiline => TextInputType.multiline,
        _ => TextInputType.text,
      },
      textInputAction: field.type == EditFieldType.multiline
          ? TextInputAction.newline
          : TextInputAction.next,
      decoration: decoration,
      validator: validate,
      onChanged: field.onChanged == null
          ? null
          : (value) => field.onChanged!(value),
    );
  }
}
