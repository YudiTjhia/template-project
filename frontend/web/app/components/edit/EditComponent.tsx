'use client';

import { useState } from 'react';
import { Box, Stack } from '@mui/material';
import PageTitle from '../page/PageTitle';
import EditActionButtons, { type EditActionButtonsProps } from './EditActionButtons';
import EditFields, { type EditFieldMetadata, type EditFieldsValues } from './EditFields';

export type EditComponentProps = {
  title: string;
  subtitle?: string;
  fields: EditFieldMetadata[];
  initialValues?: EditFieldsValues;
  error?: string;
  loading?: boolean;
  showBackButton?: boolean;
  onBack?: () => void;
  showCancel?: boolean;
  showSaveNew?: boolean;
  showUpdate?: boolean;
  showDelete?: boolean;
  onCancel?: () => void;
  onSaveNew?: (values: EditFieldsValues) => void | Promise<void>;
  onUpdate?: (values: EditFieldsValues, recordId?: string | number) => void | Promise<void>;
  onDelete?: (recordId?: string | number) => void | Promise<void>;
  actionLabels?: Pick<EditActionButtonsProps,
    'backLabel' | 'cancelLabel' | 'saveNewLabel' | 'updateLabel' | 'deleteLabel'
  >;
  actionDirection?: 'row' | 'column';
  actionAlign?: 'left' | 'center' | 'right';
  contentMaxWidth?: number | string;
  showActionButtons?: boolean;
  showSubmitButton?: boolean;
};

export default function EditComponent({
  title,
  subtitle,
  fields,
  initialValues,
  error,
  loading = false,
  showBackButton = false,
  onBack,
  showCancel = true,
  showSaveNew = false,
  showUpdate = true,
  showDelete = false,
  onCancel,
  onSaveNew,
  onUpdate,
  onDelete,
  actionLabels,
  actionDirection = 'row',
  actionAlign = 'right',
  contentMaxWidth = 720,
  showActionButtons = true,
  showSubmitButton = true,
}: EditComponentProps) {
  const [values, setValues] = useState<EditFieldsValues>(initialValues ?? {});
  const [recordId, setRecordId] = useState<string | number | undefined>(
    fields.find((field) => field.recordId !== undefined)?.recordId,
  );
  const [saving, setSaving] = useState(false);

  const runAction = async (action?: (currentValues: EditFieldsValues, id?: string | number) => void | Promise<void>) => {
    if (!action || saving || loading) return;
    setSaving(true);
    try {
      await action(values, recordId);
    } finally {
      setSaving(false);
    }
  };

  const handleSubmit = async (nextValues: EditFieldsValues, nextRecordId?: string | number) => {
    setValues(nextValues);
    if (nextRecordId !== undefined) setRecordId(nextRecordId);
    const action = showSaveNew && !showUpdate ? onSaveNew : onUpdate;
    if (!action || saving || loading) return;
    setSaving(true);
    try {
      await action(nextValues, nextRecordId ?? recordId);
    } finally {
      setSaving(false);
    }
  };

  return (
    <Box sx={{ width: '100%', maxWidth: 1200, mx: 'auto', p: { xs: 2, md: 4 } }}>
      <PageTitle
        title={title}
        subtitle={subtitle}
        showBackButton={showBackButton}
        onBack={onBack}
      />
      <Stack spacing={2} sx={{ maxWidth: contentMaxWidth, mx: 'auto' }}>
        <EditFields
          fields={fields}
          initialValues={initialValues}
          loading={loading || saving}
          error={error}
          onSubmit={handleSubmit}
          onCancel={showCancel ? onCancel : undefined}
          showSubmitButton={showSubmitButton}
        />
        {showActionButtons && <EditActionButtons
          showBack={false}
          showCancel={showCancel}
          showSaveNew={showSaveNew}
          showUpdate={showUpdate}
          showDelete={showDelete}
          onCancel={onCancel}
          onSaveNew={() => runAction(onSaveNew)}
          onUpdate={() => runAction(onUpdate)}
          onDelete={() => onDelete?.(recordId)}
          loading={loading || saving}
          disabled={loading || saving}
          direction={actionDirection}
          align={actionAlign}
          {...actionLabels}
        />}
      </Stack>
    </Box>
  );
}
