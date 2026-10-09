'use client';

import { useEffect, useState } from 'react';
import { Alert, Button, Chip, CircularProgress, Stack } from '@mui/material';
import EditComponent from '@/app/components/edit/EditComponent';
import type { EditFieldMetadata, EditFieldsValues } from '@/app/components/edit/EditFields';
import userService, { type UserPayload, type UserRecord } from '@/app/services/UserService';
import roleService from '@/app/services/RoleService';

export type UserComponentProps = {
  mode: 'create' | 'edit' | 'view';
  userId?: number;
  onSaved?: (userId?: number) => void;
  onCancel?: () => void;
};

const toValues = (user: UserRecord): EditFieldsValues => ({
  username: user.username,
  user_email: user.user_email,
  mobilephone: user.mobilephone,
  user_type: user.user_type,
  login_type: user.login_type,
  user_roles: user.user_roles ?? [],
  is_active: user.is_active,
});

export default function UserComponent({ mode, userId, onSaved, onCancel }: UserComponentProps) {
  const [user, setUser] = useState<UserRecord>();
  const [roleNames, setRoleNames] = useState<string[]>([]);
  const [loading, setLoading] = useState(mode !== 'create');
  const [error, setError] = useState<string>();

  useEffect(() => {
    let mounted = true;
    Promise.all([roleService.list(), mode !== 'create' && userId ? userService.get(userId) : Promise.resolve(undefined)])
      .then(([roles, record]) => {
        if (!mounted) return;
        setRoleNames(roles.map((role) => role.role_name));
        setUser(record);
      }).catch((caught) => {
        if (mounted) setError(caught instanceof Error ? caught.message : 'Gagal memuat user.');
      }).finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, [mode, userId]);

  if (loading) return <Stack alignItems="center" sx={{ py: 8 }}><CircularProgress /></Stack>;

  const readOnly = mode === 'view';
  const fields: EditFieldMetadata[] = [
    { id: 'username', label: 'Username', type: 'text', required: true, disabled: readOnly },
    { id: 'user_email', label: 'Email', type: 'email', required: true, disabled: readOnly },
    { id: 'mobilephone', label: 'Nomor telepon', type: 'tel', required: true, disabled: readOnly },
    { id: 'user_password', label: mode === 'create' ? 'Password' : 'Password baru', type: 'password', required: mode === 'create', disabled: readOnly, helperText: mode === 'edit' ? 'Kosongkan jika tidak ingin mengganti password.' : undefined },
    { id: 'user_type', label: 'Tipe user', type: 'select', required: true, disabled: readOnly, options: [
      { label: 'User', value: 'USER' }, { label: 'Super Admin', value: 'SUPER_ADMIN' },
    ] },
    { id: 'login_type', label: 'Tipe login', type: 'select', required: true, disabled: readOnly, options: [
      { label: 'Kasir', value: 'cashier' }, { label: 'Customer table', value: 'customer_table' },
    ] },
    { id: 'user_roles', label: 'Role', type: 'select-multiple', required: true, disabled: readOnly, options: roleNames.map((name) => ({ label: name, value: name })) },
    { id: 'is_active', label: 'Aktif', type: 'checkbox', disabled: readOnly },
  ];

  const submit = async (values: EditFieldsValues) => {
    const payload: UserPayload = {
      username: String(values.username ?? ''),
      user_email: String(values.user_email ?? ''),
      mobilephone: String(values.mobilephone ?? ''),
      user_type: String(values.user_type ?? 'USER'),
      login_type: String(values.login_type ?? 'cashier'),
      user_roles: Array.isArray(values.user_roles) ? values.user_roles.map(String) : [],
      is_active: Boolean(values.is_active),
      ...(String(values.user_password ?? '') ? { user_password: String(values.user_password) } : {}),
    };
    if (mode === 'create') {
      const result = await userService.create(payload);
      onSaved?.(result.id);
    } else if (userId) {
      await userService.update(userId, payload);
      onSaved?.(userId);
    }
  };

  return (
    <>
      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      <EditComponent
        title={mode === 'create' ? 'Tambah user' : mode === 'view' ? 'Detail user' : 'Edit user'}
        subtitle="Kelola identitas, akses, dan status user."
        fields={fields}
        initialValues={user ? toValues(user) : { user_type: 'USER', login_type: 'cashier', is_active: true }}
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
        actionLabels={{ saveNewLabel: 'Buat user', updateLabel: 'Simpan perubahan' }}
      />
    </>
  );
}
