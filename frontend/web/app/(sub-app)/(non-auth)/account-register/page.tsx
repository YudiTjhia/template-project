'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Box } from '@mui/material';
import { AccountRegisterForm, type AccountRegisterValues } from '@/app/components/account';
import { AuthPagePanel } from '@/app/components/page';
import accountService from '@/app/services/AccountService';

type RegisterResponse = { id?: number; account_id?: number };

export default function AccountRegisterPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();

  const handleRegister = async ({ name, email, password }: AccountRegisterValues) => {
    setLoading(true);
    setError(undefined);
    try {
      const response = await accountService.register<RegisterResponse>({ name, email, password });
      const accountId = response.data?.id ?? response.data?.account_id;
      if (!accountId) throw new Error('Pendaftaran berhasil dimulai, tetapi ID akun tidak diterima.');
      router.push(`/account-otp?accountId=${accountId}&email=${encodeURIComponent(email)}`);
    } catch (caughtError) {
      const message = caughtError instanceof Error ? caughtError.message : 'Pendaftaran akun gagal.';
      setError(message);
      throw caughtError;
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      component="main"
      sx={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        py: 4,
        backgroundColor: 'primary.main',
      }}
    >
      <AuthPagePanel>
        <AccountRegisterForm
          onSubmit={handleRegister}
          onLoginClick={() => router.push('/login')}
          loading={loading}
          error={error}
        />
      </AuthPagePanel>
    </Box>
  );
}
