'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Box } from '@mui/material';
import { LoginForm, type LoginValues } from '@/app/components/login';
import { AuthPagePanel } from '@/app/components/page';
import loginService from '@/app/services/LoginService';
import { getSubAppPaths, resolveSubAppType } from '@/app/helpers/subApp';

export default function LoginPage() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();

  const handleLogin = async (values: LoginValues) => {
    setLoading(true);
    setError(undefined);
    try {
      await loginService.login(values);
      router.replace(getSubAppPaths(resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE)).home);
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Login gagal.');
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
        <LoginForm
          onSubmit={handleLogin}
          onRegisterClick={() => router.push('/account-register')}
          onForgotPasswordClick={() => router.push('/login-otp')}
          loading={loading}
          error={error}
        />
      </AuthPagePanel>
    </Box>
  );
}
