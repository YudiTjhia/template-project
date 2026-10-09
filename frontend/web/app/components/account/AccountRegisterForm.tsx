'use client';

import { useState } from 'react';
import { Link as MuiLink, Stack, Typography } from '@mui/material';
import EditCenter from '../edit/EditCenter';
import type { EditFieldMetadata, EditFieldsValues } from '../edit/EditFields';

export type AccountRegisterValues = {
  name: string;
  email: string;
  address: string;
  mobilephone: string;
  password: string;
  confirmPassword: string;
  pin: string;
  confirmPin: string;
  acceptTerms: boolean;
};

export type AccountRegisterFormProps = {
  onSubmit: (values: AccountRegisterValues) => void | Promise<void>;
  onLoginClick?: () => void;
  loading?: boolean;
  error?: string;
};

const fields: EditFieldMetadata[] = [
  {
    id: 'name',
    field: 'name',
    label: 'Nama lengkap',
    type: 'text',
    placeholder: 'Masukkan nama lengkap',
    required: true,
    value: '',
  },
  {
    id: 'email',
    field: 'email',
    label: 'Email',
    type: 'email',
    placeholder: 'nama@email.com',
    required: true,
    value: '',
  },
  {
    id: 'address',
    field: 'address',
    label: 'Alamat',
    type: 'textarea',
    placeholder: 'Masukkan alamat',
    required: true,
    value: '',
  },
  {
    id: 'mobilephone',
    field: 'mobilephone',
    label: 'Nomor ponsel',
    type: 'tel',
    placeholder: 'Contoh: 081234567890',
    required: true,
    value: '',
  },
  {
    id: 'password',
    field: 'password',
    label: 'Kata sandi',
    type: 'password',
    placeholder: 'Minimal 8 karakter',
    helperText: 'Gunakan minimal 8 karakter.',
    required: true,
    value: '',
  },
  {
    id: 'confirmPassword',
    field: 'confirmPassword',
    label: 'Konfirmasi kata sandi',
    type: 'password',
    placeholder: 'Ulangi kata sandi',
    required: true,
    value: '',
  },
  {
    id: 'pin',
    field: 'pin',
    label: 'PIN',
    type: 'password',
    placeholder: '6 digit angka',
    helperText: 'Gunakan 6 digit angka.',
    required: true,
    pattern: '[0-9]{6}',
    value: '',
  },
  {
    id: 'confirmPin',
    field: 'confirmPin',
    label: 'Konfirmasi PIN',
    type: 'password',
    placeholder: 'Ulangi PIN',
    required: true,
    pattern: '[0-9]{6}',
    value: '',
  },
  {
    id: 'acceptTerms',
    field: 'acceptTerms',
    label: 'Saya menyetujui syarat dan ketentuan.',
    type: 'checkbox',
    required: true,
    value: false,
  },
];


export default function AccountRegisterForm({
  onSubmit,
  onLoginClick,
  loading = false,
  error,
}: AccountRegisterFormProps) {
  const [validationError, setValidationError] = useState<string>();

  const handleSubmit = async (formValues: EditFieldsValues) => {
    const values = formValues as AccountRegisterValues;
    if (values.password !== values.confirmPassword) {
      setValidationError('Konfirmasi kata sandi tidak cocok.');
      return;
    }
    if (!/^[0-9]{6}$/.test(values.pin)) {
      setValidationError('PIN harus terdiri dari 6 digit angka.');
      return;
    }
    if (values.pin !== values.confirmPin) {
      setValidationError('Konfirmasi PIN tidak cocok.');
      return;
    }
    if (!values.acceptTerms) {
      setValidationError('Setujui syarat dan ketentuan untuk melanjutkan.');
      return;
    }

    setValidationError(undefined);
    await onSubmit(values);
  };

  return (
    <EditCenter
      title="Buat akun"
      description="Daftar untuk mulai menggunakan layanan."
      panelSx={{ border: 0, bgcolor: 'common.white', boxShadow: 'none' }}
      fields={fields}
      loading={loading}
      error={validationError ?? error}
      onSubmit={handleSubmit}
      actions={(
        <Stack direction="row" justifyContent="center" spacing={0.5}>
          <Typography variant="body2" color="text.secondary">Sudah punya akun?</Typography>
          <MuiLink
            component="button"
            type="button"
            variant="body2"
            underline="hover"
            onClick={onLoginClick}
            disabled={!onLoginClick || loading}
          >
            Masuk
          </MuiLink>
        </Stack>
      )}
    />
  );
}
