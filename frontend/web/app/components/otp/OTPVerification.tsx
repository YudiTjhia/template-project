'use client';

import { useState, type ReactNode } from 'react';
import { Button, Stack } from '@mui/material';
import EditCenter from '@/app/components/edit/EditCenter';
import type { EditFieldMetadata, EditFieldsValues } from '@/app/components/edit/EditFields';

export type OTPVerificationProps = {
  title: string;
  description: string;
  helperText?: string;
  submitLabel?: string;
  resendLabel?: string;
  onVerify: (otp: string) => void | Promise<void>;
  onResend?: () => void | Promise<void>;
  afterActions?: ReactNode;
  loading?: boolean;
  resending?: boolean;
  error?: string;
};

const otpFields: EditFieldMetadata[] = [{
  id: 'otp',
  field: 'otp',
  label: 'Kode OTP',
  type: 'text',
  placeholder: 'Masukkan 6 digit OTP',
  helperText: 'Masukkan kode enam angka yang Anda terima.',
  required: true,
  pattern: '^[0-9]{6}$',
  value: '',
}];

export default function OTPVerification({
  title,
  description,
  helperText,
  submitLabel = 'Verifikasi OTP',
  resendLabel = 'Kirim ulang OTP',
  onVerify,
  onResend,
  afterActions,
  loading = false,
  resending = false,
  error,
}: OTPVerificationProps) {
  const [actionError, setActionError] = useState<string>();

  const handleVerify = async (values: EditFieldsValues) => {
    const otp = String(values.otp ?? '').trim();
    if (!/^\d{6}$/.test(otp)) throw new Error('OTP harus terdiri dari 6 angka.');
    setActionError(undefined);
    await onVerify(otp);
  };

  const handleResend = async () => {
    if (!onResend) return;
    setActionError(undefined);
    try {
      await onResend();
    } catch (caughtError) {
      setActionError(caughtError instanceof Error ? caughtError.message : 'Gagal mengirim ulang OTP.');
    }
  };

  return (
    <EditCenter
      title={title}
      description={description}
      fields={helperText ? otpFields.map((field) => ({ ...field, helperText })) : otpFields}
      submitLabel={submitLabel}
      loading={loading || resending}
      error={actionError ?? error}
      onSubmit={handleVerify}
      actions={(
        <Stack spacing={1.5} alignItems="center">
          {onResend && (
            <Button onClick={handleResend} disabled={loading || resending}>
              {resending ? 'Mengirim ulang…' : resendLabel}
            </Button>
          )}
          {afterActions}
        </Stack>
      )}
    />
  );
}
