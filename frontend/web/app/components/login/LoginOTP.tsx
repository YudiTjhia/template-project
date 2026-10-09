'use client';

import { useState } from 'react';
import { Button, Link as MuiLink, Stack, Typography } from '@mui/material';
import EditCenter from '../edit/EditCenter';
import type { EditFieldMetadata, EditFieldsValues } from '../edit/EditFields';
import { OTPVerification } from '@/app/components/otp';

export type LoginOTPProps = {
  onRequestOTP: (identifier: string) => void | Promise<void>;
  onVerifyOTP: (identifier: string, otp: string) => void | Promise<void>;
  onBackToLogin?: () => void;
  loading?: boolean;
  resending?: boolean;
  error?: string;
};

const identifierFields: EditFieldMetadata[] = [{
  id: 'identifier',
  field: 'identifier',
  label: 'Email atau username',
  type: 'text',
  placeholder: 'Masukkan email atau username',
  required: true,
  value: '',
}];

export default function LoginOTP({
  onRequestOTP,
  onVerifyOTP,
  onBackToLogin,
  loading = false,
  resending = false,
  error,
}: LoginOTPProps) {
  const [identifier, setIdentifier] = useState('');
  const [actionError, setActionError] = useState<string>();

  const handleRequest = async (values: EditFieldsValues) => {
    const nextIdentifier = String(values.identifier ?? '').trim();
    if (!nextIdentifier) throw new Error('Email atau username wajib diisi.');
    await onRequestOTP(nextIdentifier);
    setIdentifier(nextIdentifier);
    setActionError(undefined);
  };

  const handleResend = async () => {
    if (!identifier) return;
    setActionError(undefined);
    try {
      await onRequestOTP(identifier);
    } catch (caughtError) {
      setActionError(caughtError instanceof Error ? caughtError.message : 'Gagal mengirim ulang OTP.');
      throw caughtError;
    }
  };

  if (identifier) {
    return (
      <OTPVerification
        title="Verifikasi OTP login"
        description={`Masukkan kode OTP untuk ${identifier}.`}
        helperText="Kode OTP dikirim ke email akun Anda."
        onVerify={(otp) => onVerifyOTP(identifier, otp)}
        onResend={handleResend}
        resending={resending}
        loading={loading}
        error={actionError ?? error}
        afterActions={(
          <>
            <Button
              size="small"
              onClick={() => { setIdentifier(''); setActionError(undefined); }}
              disabled={loading || resending}
            >
              Ganti email atau username
            </Button>
            {onBackToLogin && (
              <MuiLink component="button" type="button" variant="body2" onClick={onBackToLogin} disabled={loading || resending}>
                Kembali ke login
              </MuiLink>
            )}
          </>
        )}
      />
    );
  }

  return (
    <EditCenter
      title="Masuk dengan OTP"
      description="Masukkan email atau username untuk menerima kode OTP."
      fields={identifierFields}
      submitLabel="Kirim OTP"
      loading={loading}
      error={error}
      onSubmit={handleRequest}
      actions={onBackToLogin ? (
        <Stack direction="row" justifyContent="center" spacing={0.5}>
          <Typography variant="body2" color="text.secondary">Ingin menggunakan kata sandi?</Typography>
          <MuiLink component="button" type="button" variant="body2" onClick={onBackToLogin} disabled={loading}>
            Kembali ke login
          </MuiLink>
        </Stack>
      ) : undefined}
    />
  );
}
