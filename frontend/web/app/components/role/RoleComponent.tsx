'use client';

import { useEffect, useState } from 'react';
import { Alert, Button, Chip, CircularProgress, Stack } from '@mui/material';
import EditComponent from '@/app/components/edit/EditComponent';
import type { EditFieldMetadata, EditFieldsValues } from '@/app/components/edit/EditFields';
import { RoleActionLookupDialog } from '@/app/components/role_action';
import roleService, { type RoleRecord } from '@/app/services/RoleService';

export type RoleComponentProps = {
  mode: 'create' | 'edit' | 'view';
  roleId?: number;
  onSaved?: (roleId?: number) => void;
  onCancel?: () => void;
};

export default function RoleComponent({ mode, roleId, onSaved, onCancel }: RoleComponentProps) {
  const [role, setRole] = useState<RoleRecord>();
  const [actions, setActions] = useState<string[]>([]);
  const [lookupOpen, setLookupOpen] = useState(false);
  const [loading, setLoading] = useState(mode !== 'create');
  const [error, setError] = useState<string>();

  useEffect(() => {
    if (mode === 'create') return;
    let mounted = true;
    roleService.get(roleId!).then((record) => {
      if (!mounted) return;
      setRole(record);
      setActions(record.actions ?? []);
    }).catch((caught) => {
      if (mounted) setError(caught instanceof Error ? caught.message : 'Gagal memuat role.');
    }).finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, [mode, roleId]);

  if (loading) return <Stack alignItems="center" sx={{ py: 8 }}><CircularProgress /></Stack>;

  const readOnly = mode === 'view';
  const fields: EditFieldMetadata[] = [
    { id: 'role_name', label: 'Nama role', type: 'text', required: true, disabled: readOnly },
    { id: 'actions', label: 'Kode action', type: 'textarea', disabled: true, helperText: actions.length ? actions.join(', ') : 'Belum ada action dipilih.' },
  ];

  const submit = async (values: EditFieldsValues) => {
    const payload = { role_name: String(values.role_name ?? ''), actions };
    if (mode === 'create') {
      const result = await roleService.create(payload);
      onSaved?.(result.id);
    } else if (roleId) {
      await roleService.update(roleId, payload);
      onSaved?.(roleId);
    }
  };

  return (
    <>
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      <EditComponent
        title={mode === 'create' ? 'Tambah role' : mode === 'view' ? 'Detail role' : 'Edit role'}
        subtitle="Atur nama role dan daftar action yang dimiliki."
        fields={fields}
        initialValues={role ? { role_name: role.role_name, actions: (role.actions ?? []).join(', ') } : { role_name: '', actions: '' }}
        showBackButton
        onBack={onCancel}
        showCancel={mode !== 'view'}
        onCancel={onCancel}
        showSaveNew={mode === 'create'}
        showUpdate={mode === 'edit'}
        showDelete={false}
        showActionButtons={false}
        showSubmitButton={!readOnly}
        onSaveNew={submit}
        onUpdate={submit}
        actionLabels={{ saveNewLabel: 'Buat role', updateLabel: 'Simpan perubahan' }}
      />
      <Stack direction="row" spacing={1} useFlexGap flexWrap="wrap" sx={{ maxWidth: 720, mx: 'auto', mt: 2 }}>
        {!readOnly && <Button variant="outlined" onClick={() => setLookupOpen(true)}>Pilih role action</Button>}
        {actions.map((action) => <Chip key={action} label={action} onDelete={!readOnly ? () => setActions((current) => current.filter((item) => item !== action)) : undefined} />)}
      </Stack>
      <RoleActionLookupDialog
        open={lookupOpen}
        selectedValues={actions}
        onClose={() => setLookupOpen(false)}
        onApply={(selected) => { setActions(selected.map((action) => action.value)); setLookupOpen(false); }}
      />
    </>
  );
}
