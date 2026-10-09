'use client';

import { AppRouterCacheProvider } from '@mui/material-nextjs/v15-appRouter';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { useEffect, useState } from 'react';
import { InitialSplashScreen } from '@/app/components/page';
import theme from './themes/theme';
import { ConfirmBoxProvider } from '@/app/components/confirmbox';
import { MsgBoxProvider } from '@/app/components/msgbox';

export default function Providers({ children }: { children: React.ReactNode }) {
  const [showInitialSplash, setShowInitialSplash] = useState(true);

  useEffect(() => {
    const timer = window.setTimeout(() => setShowInitialSplash(false), 2500);
    return () => window.clearTimeout(timer);
  }, []);

  return (
    <AppRouterCacheProvider>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <MsgBoxProvider>
          <ConfirmBoxProvider>
            {children}
            {showInitialSplash && <InitialSplashScreen />}
          </ConfirmBoxProvider>
        </MsgBoxProvider>
      </ThemeProvider>
    </AppRouterCacheProvider>
  );
}
