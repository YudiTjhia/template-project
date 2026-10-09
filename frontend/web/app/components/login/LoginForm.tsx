'use client';

import { Link as MuiLink, Stack, Typography } from '@mui/material';
import EditCenter from '../edit/EditCenter';
import type { EditFieldMetadata, EditFieldsValues } from '../edit/EditFields';

export type LoginValues = {
  email: string;
  password: string;
};

export type LoginFormProps = {
  onSubmit: (values: LoginValues) => void | Promise<void>;
  onRegisterClick?: () => void;
  onForgotPasswordClick?: () => void;
  loading?: boolean;
  error?: string;
};

const fields: EditFieldMetadata[] = [
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
    id: 'password',
    field: 'password',
    label: 'Kata sandi',
    type: 'password',
    placeholder: 'Masukkan kata sandi',
    required: true,
    value: '',
  },
];

export default function LoginForm({
  onSubmit,
  onRegisterClick,
  onForgotPasswordClick,
  loading = false,
  error,
}: LoginFormProps) {
  const handleSubmit = async (formValues: EditFieldsValues) => {
    await onSubmit(formValues as LoginValues);
  };

  return (
    <EditCenter
      title="Masuk"
      description="Masuk ke akun Anda untuk melanjutkan."
      fields={fields}
      loading={loading}
      error={error}
      onSubmit={handleSubmit}
      actions={(
        <Stack spacing={1} alignItems="center">
          {onForgotPasswordClick && (
            <MuiLink component="button" type="button" variant="body2" onClick={onForgotPasswordClick} disabled={loading}>
              Lupa kata sandi?
            </MuiLink>
          )}
          {onRegisterClick && (
            <Stack direction="row" justifyContent="center" spacing={0.5}>
              <Typography variant="body2" color="text.secondary">Belum punya akun?</Typography>
              <MuiLink component="button" type="button" variant="body2" underline="hover" onClick={onRegisterClick} disabled={loading}>
                Daftar
              </MuiLink>
            </Stack>
          )}
        </Stack>
      )}
    />
  );
}
