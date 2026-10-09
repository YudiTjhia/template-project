'use client';

import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { AppBar, Box, Button, CircularProgress, Stack, Toolbar, Typography } from '@mui/material';
import { useRouter } from 'next/navigation';
import loginService from '@/app/services/LoginService';
import { getSubAppPaths, resolveSubAppType, type SubAppType } from '@/app/helpers/subApp';

export type PublicLayoutProps = {
  children: ReactNode;
  title?: string;
  appName?: string;
  loginHref?: string;
  registerHref?: string;
  showHeader?: boolean;
  showLogin?: boolean;
  showRegister?: boolean;
  subAppType?: SubAppType;
};

export default function PublicLayout({
  children,
  title,
  appName = 'Application',
  loginHref = '/login',
  registerHref = '/account-register',
  showHeader = true,
  showLogin = true,
  showRegister = true,
  subAppType = resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE),
}: PublicLayoutProps) {
  const router = useRouter();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    let active = true;
    const validate = async () => {
      const token = loginService.getSessionToken();
      if (!token) {
        if (active) setChecking(false);
        return;
      }

      try {
        await loginService.validateSession(token, subAppType);
        if (active) router.replace(getSubAppPaths(subAppType).home);
      } catch {
        loginService.clearSessionToken();
        if (active) setChecking(false);
      }
    };

    void validate();
    return () => {
      active = false;
    };
  }, [router, subAppType]);

  if (checking) {
    return (
      <Stack minHeight="100vh" alignItems="center" justifyContent="center" spacing={2}>
        <CircularProgress size={28} />
        <Typography color="text.secondary">Memeriksa sesi…</Typography>
      </Stack>
    );
  }

  return (
    <Box sx={{ minHeight: '100vh', display: 'flex', flexDirection: 'column', bgcolor: 'background.default' }}>
      {showHeader && (
        <AppBar position="static" color="inherit" elevation={1}>
          <Toolbar>
            <Typography variant="h6" component="div" sx={{ flexGrow: 1, fontWeight: 600 }}>
              {title ?? appName}
            </Typography>
            {showLogin && <Button onClick={() => router.push(loginHref)}>Login</Button>}
            {showRegister && (
              <Button variant="contained" onClick={() => router.push(registerHref)} sx={{ ml: 1 }}>
                Register
              </Button>
            )}
          </Toolbar>
        </AppBar>
      )}
      <Box component="main" sx={{ flex: 1, width: '100%' }}>
        {children}
      </Box>
    </Box>
  );
}
