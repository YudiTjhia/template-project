'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Checkbox,
  FormControl,
  FormControlLabel,
  FormHelperText,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  TextField,
  Typography,
} from '@mui/material';

export type EditFieldType =
  | 'text'
  | 'email'
  | 'password'
  | 'number'
  | 'tel'
  | 'url'
  | 'date'
  | 'datetime-local'
  | 'time'
  | 'textarea'
  | 'select'
  | 'checkbox';

export type EditFieldOption = {
  id?: string | number;
  label: string;
  value: string | number;
};

export type EditFieldMetadata = {
  /** Stable identifier used as the field name in form values. */
  id: string;
  /** Optional data record identifier; submitted separately from field values. */
  recordId?: string | number;
  /** Data key for API mapping. Defaults to id. */
  field?: string;
  label: string;
  type: EditFieldType;
  value?: unknown;
  placeholder?: string;
  helperText?: string;
  required?: boolean;
  disabled?: boolean;
  readOnly?: boolean;
  min?: number | string;
  max?: number | string;
  step?: number | string;
  rows?: number;
  options?: EditFieldOption[];
  /** Basic HTML input validation pattern. */
  pattern?: string;
};

export type EditFieldsValues = Record<string, unknown>;

export type EditFieldsProps = {
  fields: EditFieldMetadata[];
  title?: string;
  description?: string;
  submitLabel?: string;
  cancelLabel?: string;
  initialValues?: EditFieldsValues;
  loading?: boolean;
  error?: string;
  onSubmit: (values: EditFieldsValues, recordId?: string | number) => void | Promise<void>;
  onCancel?: () => void;
};

function getInitialValues(fields: EditFieldMetadata[], suppliedValues: EditFieldsValues = {}) {
  return fields.reduce<EditFieldsValues>((values, metadata) => {
    values[metadata.id] = suppliedValues[metadata.id] ?? metadata.value ?? (metadata.type === 'checkbox' ? false : '');
    return values;
  }, {});
}

export default function EditFields({
  fields,
  title,
  description,
  submitLabel = 'Simpan perubahan',
  cancelLabel = 'Batal',
  initialValues,
  loading = false,
  error,
  onSubmit,
  onCancel,
}: EditFieldsProps) {
  const initialRecordId = fields.find((field) => field.recordId !== undefined)?.recordId;
  const [values, setValues] = useState<EditFieldsValues>(() => getInitialValues(fields, initialValues));
  const [submitError, setSubmitError] = useState<string>();
  const inputIdPrefix = useMemo(() => `edit-fields-${Math.random().toString(36).slice(2, 9)}`, []);

  useEffect(() => {
    setValues(getInitialValues(fields, initialValues));
  }, [fields, initialValues]);

  const setValue = (id: string, value: unknown) => {
    setValues((current) => ({ ...current, [id]: value }));
  };

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSubmitError(undefined);
    try {
      await onSubmit(values, initialRecordId);
    } catch (caughtError) {
      setSubmitError(caughtError instanceof Error ? caughtError.message : 'Gagal menyimpan perubahan.');
    }
  };

  const renderField = (metadata: EditFieldMetadata) => {
    const inputId = `${inputIdPrefix}-${metadata.id}`;
    const value = values[metadata.id];

    if (metadata.type === 'checkbox') {
      return (
        <FormControl key={metadata.id} required={metadata.required} disabled={metadata.disabled} error={Boolean(metadata.helperText)}>
          <FormControlLabel
            control={
              <Checkbox
                id={inputId}
                checked={Boolean(value)}
                required={metadata.required}
                disabled={metadata.disabled}
                onChange={(event) => setValue(metadata.id, event.target.checked)}
              />
            }
            label={metadata.label}
          />
          {metadata.helperText && <FormHelperText>{metadata.helperText}</FormHelperText>}
        </FormControl>
      );
    }

    if (metadata.type === 'select') {
      const labelId = `${inputId}-label`;
      return (
        <FormControl key={metadata.id} fullWidth required={metadata.required} disabled={metadata.disabled}>
          <InputLabel id={labelId}>{metadata.label}</InputLabel>
          <Select
            id={inputId}
            labelId={labelId}
            label={metadata.label}
            value={value ?? ''}
            onChange={(event) => setValue(metadata.id, event.target.value)}
          >
            {(metadata.options ?? []).map((option) => (
              <MenuItem key={option.id ?? String(option.value)} value={option.value}>{option.label}</MenuItem>
            ))}
          </Select>
          {metadata.helperText && <FormHelperText>{metadata.helperText}</FormHelperText>}
        </FormControl>
      );
    }

    return (
      <TextField
        key={metadata.id}
        id={inputId}
        name={metadata.field ?? metadata.id}
        label={metadata.label}
        type={metadata.type === 'textarea' ? undefined : metadata.type}
        value={value ?? ''}
        placeholder={metadata.placeholder}
        helperText={metadata.helperText}
        required={metadata.required}
        disabled={metadata.disabled}
        slotProps={{
          htmlInput: {
            readOnly: metadata.readOnly,
            min: metadata.min,
            max: metadata.max,
            step: metadata.step,
            pattern: metadata.pattern,
          },
        }}
        multiline={metadata.type === 'textarea'}
        minRows={metadata.type === 'textarea' ? metadata.rows ?? 3 : undefined}
        fullWidth
        onChange={(event) => setValue(metadata.id, metadata.type === 'number' ? event.target.value : event.target.value)}
      />
    );
  };

  return (
    <Paper variant="outlined" sx={{ width: '100%', maxWidth: 720, mx: 'auto', p: { xs: 2, sm: 3 }, borderRadius: 2 }}>
      <Box component="form" onSubmit={handleSubmit} noValidate>
        {(title || description) && (
          <Stack spacing={0.5} sx={{ mb: 3 }}>
            {title && <Typography variant="h5" component="h1" fontWeight={700}>{title}</Typography>}
            {description && <Typography color="text.secondary">{description}</Typography>}
          </Stack>
        )}

        <Stack spacing={2.5}>
          {error && <Alert severity="error">{error}</Alert>}
          {submitError && <Alert severity="error">{submitError}</Alert>}
          {fields.map(renderField)}
          <Stack direction="row" justifyContent="flex-end" spacing={1.5}>
            {onCancel && <Button onClick={onCancel} disabled={loading}>{cancelLabel}</Button>}
            <Button type="submit" variant="contained" disabled={loading}>
              {loading ? 'Menyimpan…' : submitLabel}
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Paper>
  );
}
