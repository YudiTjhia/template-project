'use client';

import { Suspense, useState } from 'react';
import { CircularProgress, Stack } from '@mui/material';
import { useRouter } from 'next/navigation';
import { LoginOTP } from '@/app/components/login';
import PageContainer from '@/app/components/page/PageContainer';
import loginService from '@/app/services/LoginService';

function LoginOTPContent() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [resending, setResending] = useState(false);
  const [error, setError] = useState<string>();

  const handleRequestOTP = async (identifier: string) => {
    setLoading(true);
    setError(undefined);
    try {
      await loginService.requestLoginOTP(identifier);
    } catch (caughtError) {
      const message = caughtError instanceof Error ? caughtError.message : 'Permintaan OTP gagal.';
      setError(message);
      throw caughtError;
    } finally {
      setLoading(false);
      setResending(false);
    }
  };

  const handleVerifyOTP = async (identifier: string, otp: string) => {
    setLoading(true);
    setError(undefined);
    try {
      await loginService.loginOTP(identifier, otp);
      router.push('/');
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Verifikasi OTP gagal.');
    } finally {
      setLoading(false);
    }
  };

  const handleResendOTP = async (identifier: string) => {
    setResending(true);
    await handleRequestOTP(identifier);
  };

  return (
    <PageContainer maxWidth={520}>
      <LoginOTP
        onRequestOTP={handleRequestOTP}
        onVerifyOTP={handleVerifyOTP}
        onBackToLogin={() => router.push('/login')}
        loading={loading}
        resending={resending}
        error={error}
      />
    </PageContainer>
  );
}

export default function LoginOTPPage() {
  return (
    <Suspense fallback={<PageContainer maxWidth={520}><Stack alignItems="center"><CircularProgress /></Stack></PageContainer>}>
      <LoginOTPContent />
    </Suspense>
  );
}
