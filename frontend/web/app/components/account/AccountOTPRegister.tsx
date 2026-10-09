'use client';

import { useState } from 'react';
import { Button, Link as MuiLink, Stack, Typography } from '@mui/material';
import EditCenter from '../edit/EditCenter';
import type { EditFieldMetadata, EditFieldsValues } from '../edit/EditFields';

export type AccountOTPRegisterProps = {
  accountId: number;
  email?: string;
  onVerify: (otp: string) => void | Promise<void>;
  onResend?: (accountId: number) => void | Promise<void>;
  onLoginClick?: () => void;
  loading?: boolean;
  resending?: boolean;
  error?: string;
};

const fields: EditFieldMetadata[] = [
  {
    id: 'otp',
    field: 'otp',
    label: 'Kode OTP',
    type: 'text',
    placeholder: 'Masukkan 6 digit OTP',
    helperText: 'Masukkan kode yang dikirim ke email Anda.',
    required: true,
    pattern: '^[0-9]{6}$',
    value: '',
  },
];

export default function AccountOTPRegister({
  accountId,
  email,
  onVerify,
  onResend,
  onLoginClick,
  loading = false,
  resending = false,
  error,
}: AccountOTPRegisterProps) {
  const [actionError, setActionError] = useState<string>();

  const handleVerify = async (values: EditFieldsValues) => {
    const otp = String(values.otp ?? '').trim();
    if (!/^\d{6}$/.test(otp)) {
      const message = 'OTP harus terdiri dari 6 angka.';
      throw new Error(message);
    }
    setActionError(undefined);
    await onVerify(otp);
  };

  const handleResend = async () => {
    if (!onResend) return;
    setActionError(undefined);
    try {
      await onResend(accountId);
    } catch (caughtError) {
      setActionError(caughtError instanceof Error ? caughtError.message : 'Gagal mengirim ulang OTP.');
    }
  };

  return (
    <EditCenter
      title="Verifikasi akun"
      description={email ? `Masukkan kode OTP yang dikirim ke ${email}.` : 'Masukkan kode OTP yang dikirim ke email Anda.'}
      fields={fields}
      submitLabel="Verifikasi OTP"
      loading={loading || resending}
      error={actionError ?? error}
      onSubmit={handleVerify}
      actions={(
        <Stack spacing={1.5} alignItems="center">
          {onResend && (
            <Button onClick={handleResend} disabled={loading || resending}>
              {resending ? 'Mengirim ulang…' : 'Kirim ulang OTP'}
            </Button>
          )}
          {onLoginClick && (
            <Stack direction="row" spacing={0.5}>
              <Typography variant="body2" color="text.secondary">Sudah verifikasi?</Typography>
              <MuiLink component="button" type="button" variant="body2" onClick={onLoginClick} disabled={loading || resending}>
                Masuk
              </MuiLink>
            </Stack>
          )}
        </Stack>
      )}
    />
  );
}
