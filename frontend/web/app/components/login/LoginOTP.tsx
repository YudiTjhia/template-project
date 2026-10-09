'use client';

import { useState } from 'react';
import { Button, Link as MuiLink, Stack, Typography } from '@mui/material';
import EditCenter from '../edit/EditCenter';
import type { EditFieldMetadata, EditFieldsValues } from '../edit/EditFields';

export type LoginOTPProps = {
  onRequestOTP: (identifier: string) => void | Promise<void>;
  onVerifyOTP: (identifier: string, otp: string) => void | Promise<void>;
  onBackToLogin?: () => void;
  loading?: boolean;
  resending?: boolean;
  error?: string;
};

type LoginOTPStep = 'identifier' | 'otp';

const identifierFields: EditFieldMetadata[] = [
  {
    id: 'identifier',
    field: 'identifier',
    label: 'Email atau username',
    type: 'text',
    placeholder: 'Masukkan email atau username',
    required: true,
    value: '',
  },
];

const otpFields: EditFieldMetadata[] = [
  {
    id: 'otp',
    field: 'otp',
    label: 'Kode OTP',
    type: 'text',
    placeholder: 'Masukkan 6 digit OTP',
    helperText: 'Kode OTP dikirim ke email akun Anda.',
    required: true,
    pattern: '^[0-9]{6}$',
    value: '',
  },
];

export default function LoginOTP({
  onRequestOTP,
  onVerifyOTP,
  onBackToLogin,
  loading = false,
  resending = false,
  error,
}: LoginOTPProps) {
  const [step, setStep] = useState<LoginOTPStep>('identifier');
  const [identifier, setIdentifier] = useState('');
  const [actionError, setActionError] = useState<string>();

  const handleRequest = async (values: EditFieldsValues) => {
    const nextIdentifier = String(values.identifier ?? '').trim();
    if (!nextIdentifier) throw new Error('Email atau username wajib diisi.');
    await onRequestOTP(nextIdentifier);
    setIdentifier(nextIdentifier);
    setActionError(undefined);
    setStep('otp');
  };

  const handleVerify = async (values: EditFieldsValues) => {
    const otp = String(values.otp ?? '').trim();
    if (!/^\d{6}$/.test(otp)) {
      const message = 'OTP harus terdiri dari 6 angka.';
      throw new Error(message);
    }
    setActionError(undefined);
    await onVerifyOTP(identifier, otp);
  };

  const handleResend = async () => {
    setActionError(undefined);
    try {
      await onRequestOTP(identifier);
    } catch (caughtError) {
      setActionError(caughtError instanceof Error ? caughtError.message : 'Gagal mengirim ulang OTP.');
    }
  };

  return step === 'identifier' ? (
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
  ) : (
    <EditCenter
      title="Verifikasi OTP login"
      description={`Masukkan kode OTP untuk ${identifier}.`}
      fields={otpFields}
      submitLabel="Verifikasi OTP"
      loading={loading || resending}
      error={actionError ?? error}
      onSubmit={handleVerify}
      actions={(
        <Stack spacing={1} alignItems="center">
          <Button onClick={handleResend} disabled={loading || resending}>
            {resending ? 'Mengirim ulang…' : 'Kirim ulang OTP'}
          </Button>
          <Button onClick={() => setStep('identifier')} disabled={loading || resending}>
            Ganti email atau username
          </Button>
          {onBackToLogin && (
            <MuiLink component="button" type="button" variant="body2" onClick={onBackToLogin} disabled={loading || resending}>
              Kembali ke login
            </MuiLink>
          )}
        </Stack>
      )}
    />
  );
}
