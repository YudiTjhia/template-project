'use client';

import { Link as MuiLink, Stack, Typography } from '@mui/material';
import { OTPVerification } from '@/app/components/otp';

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
  return (
    <OTPVerification
      title="Verifikasi akun"
      description={email ? `Masukkan kode OTP yang dikirim ke ${email}.` : 'Masukkan kode OTP yang dikirim ke email Anda.'}
      helperText="Masukkan kode yang dikirim ke email Anda."
      onVerify={onVerify}
      onResend={onResend ? () => onResend(accountId) : undefined}
      loading={loading}
      resending={resending}
      error={error}
      afterActions={onLoginClick ? (
        <Stack direction="row" spacing={0.5}>
          <Typography variant="body2" color="text.secondary">Sudah verifikasi?</Typography>
          <MuiLink component="button" type="button" variant="body2" onClick={onLoginClick} disabled={loading || resending}>
            Masuk
          </MuiLink>
        </Stack>
      ) : undefined}
    />
  );
}
