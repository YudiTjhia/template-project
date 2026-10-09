'use client';

import { Suspense, useState } from 'react';
import { Alert, CircularProgress, Stack } from '@mui/material';
import { useRouter, useSearchParams } from 'next/navigation';
import { AccountOTPRegister } from '@/app/components/account';
import PageContainer from '@/app/components/page/PageContainer';
import accountService from '@/app/services/AccountService';

function AccountOTPContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const accountId = Number(searchParams.get('accountId'));
  const email = searchParams.get('email') ?? undefined;
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [error, setError] = useState<string>();

  if (!Number.isInteger(accountId) || accountId <= 0) {
    return (
      <PageContainer maxWidth={520}>
        <Alert severity="warning">ID akun tidak ditemukan. Silakan daftar kembali untuk menerima OTP.</Alert>
      </PageContainer>
    );
  }

  const handleVerify = async (otp: string) => {
    setLoading(true);
    setError(undefined);
    try {
      await accountService.verifikasiOTP(otp);
      router.push('/login');
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Verifikasi OTP gagal.');
    } finally {
      setLoading(false);
    }
  };

  const handleResend = async (id: number) => {
    setResending(true);
    setError(undefined);
    try {
      await accountService.resendOTP(id);
    } catch (caughtError) {
      const message = caughtError instanceof Error ? caughtError.message : 'Gagal mengirim ulang OTP.';
      setError(message);
      throw caughtError;
    } finally {
      setResending(false);
    }
  };

  return (
    <PageContainer maxWidth={520}>
      <AccountOTPRegister
        accountId={accountId}
        email={email}
        onVerify={handleVerify}
        onResend={handleResend}
        onLoginClick={() => router.push('/login')}
        loading={loading}
        resending={resending}
        error={error}
      />
    </PageContainer>
  );
}

export default function AccountOTPPage() {
  return (
    <Suspense fallback={<PageContainer maxWidth={520}><Stack alignItems="center"><CircularProgress /></Stack></PageContainer>}>
      <AccountOTPContent />
    </Suspense>
  );
}
