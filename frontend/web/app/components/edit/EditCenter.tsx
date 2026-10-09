'use client';

import { Box, Paper, Stack } from '@mui/material';
import type { SxProps, Theme } from '@mui/material/styles';
import EditFields, { type EditFieldMetadata, type EditFieldsValues } from './EditFields';

export type EditCenterProps = {
  fields: EditFieldMetadata[];
  title?: string;
  description?: string;
  submitLabel?: string;
  cancelLabel?: string;
  initialValues?: EditFieldsValues;
  loading?: boolean;
  error?: string;
  maxWidth?: number | string;
  panelSx?: SxProps<Theme>;
  actions?: React.ReactNode;
  onSubmit: (values: EditFieldsValues, recordId?: string | number) => void | Promise<void>;
  onCancel?: () => void;
};

export default function EditCenter({
  fields,
  title,
  description,
  submitLabel,
  cancelLabel,
  initialValues,
  loading = false,
  error,
  maxWidth = 720,
  panelSx,
  actions,
  onSubmit,
  onCancel,
}: EditCenterProps) {
  return (
    <Box sx={{ width: '100%', display: 'flex', justifyContent: 'center' }}>
      <Paper        
        sx={{ width: '100%', maxWidth, p: { xs: 2, sm: 3 }, borderRadius: 2, ...panelSx }}
      >
        <Stack spacing={2.5}>
          <EditFields
            fields={fields}
            title={title}
            description={description}
            submitLabel={submitLabel}
            cancelLabel={cancelLabel}
            initialValues={initialValues}
            loading={loading}
            error={error}
            onSubmit={onSubmit}
            onCancel={onCancel}
          />
          {actions}
        </Stack>
      </Paper>
    </Box>
  );
}
