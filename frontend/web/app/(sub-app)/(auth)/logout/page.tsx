'use client';

import { useEffect, useRef } from 'react';
import { CircularProgress, Stack, Typography } from '@mui/material';
import { useRouter } from 'next/navigation';
import PageContainer from '@/app/components/page/PageContainer';
import loginService from '@/app/services/LoginService';

export default function LogoutPage() {
  const router = useRouter();
  const started = useRef(false);

  useEffect(() => {
    if (started.current) return;
    started.current = true;

    void loginService.logoutCurrentSession().catch(() => undefined).finally(() => {
      router.replace('/');
    });
  }, [router]);

  return (
    <PageContainer maxWidth={480}>
      <Stack alignItems="center" spacing={2}>
        <CircularProgress size={28} />
        <Typography color="text.secondary">Keluar dari sesi…</Typography>
      </Stack>
    </PageContainer>
  );
}
