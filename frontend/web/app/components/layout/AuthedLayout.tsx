'use client';

import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import { AppBar, Box, CircularProgress, Stack, Toolbar, Typography } from '@mui/material';
import { usePathname, useRouter } from 'next/navigation';
import { NavigationMenu } from '@/app/components/navigation';
import loginService from '@/app/services/LoginService';
import { getSubAppPaths, resolveSubAppType, type SubAppType } from '@/app/helpers/subApp';

export type AuthedLayoutProps = {
  children: ReactNode;
  title?: string;
  appName?: string;
  onLogout?: () => void;
  showHeader?: boolean;
  showNavigation?: boolean;
  navigationSide?: 'left' | 'right';
  subAppType?: SubAppType;
};

export default function AuthedLayout({
  children,
  title,
  appName = 'Application',
  onLogout,
  showHeader = true,
  showNavigation = true,
  navigationSide = 'left',
  subAppType = resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE),
}: AuthedLayoutProps) {
  const pathname = usePathname();
  const router = useRouter();
  const [authorized, setAuthorized] = useState(pathname === '/logout');

  useEffect(() => {
    if (pathname === '/logout') {
      setAuthorized(true);
      return;
    }

    setAuthorized(false);
    let active = true;
    const validate = async () => {
      const token = loginService.getSessionToken();
      if (!token) {
        if (active) router.replace(getSubAppPaths(subAppType).login);
        return;
      }

      try {
        await loginService.validateSession(token, subAppType);
        if (active) setAuthorized(true);
      } catch {
        loginService.clearSessionToken();
        if (active) router.replace(getSubAppPaths(subAppType).login);
      }
    };

    void validate();
    return () => {
      active = false;
    };
  }, [pathname, router, subAppType]);

  if (!authorized) {
    return (
      <Stack minHeight="100vh" alignItems="center" justifyContent="center" spacing={2}>
        <CircularProgress size={28} />
        <Typography color="text.secondary">Memeriksa sesi…</Typography>
      </Stack>
    );
  }

  const navigation = showNavigation ? (
    <NavigationMenu side={navigationSide} onLogout={onLogout} />
  ) : null;

  return (
    <Box sx={{ minHeight: '100vh', display: 'flex', bgcolor: 'background.default' }}>
      {navigationSide === 'left' && navigation}
      <Box sx={{ flex: 1, minWidth: 0, display: 'flex', flexDirection: 'column' }}>
        {showHeader && (
          <AppBar position="static" color="inherit" elevation={1}>
            <Toolbar>
              <Typography variant="h6" component="div" sx={{ flexGrow: 1, fontWeight: 600 }}>
                {title ?? appName}
              </Typography>
            </Toolbar>
          </AppBar>
        )}
        <Box component="main" sx={{ flex: 1, width: '100%' }}>
          {children}
        </Box>
      </Box>
      {navigationSide === 'right' && navigation}
    </Box>
  );
}
