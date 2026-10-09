'use client';

import { AppRouterCacheProvider } from '@mui/material-nextjs/v15-appRouter';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import theme from './themes/theme';
import { ConfirmBoxProvider } from '@/app/components/confirmbox';
import { MsgBoxProvider } from '@/app/components/msgbox';

export default function Providers({ children }: { children: React.ReactNode }) {
  return (
    <AppRouterCacheProvider>
      <ThemeProvider theme={theme}>
        <CssBaseline />
        <MsgBoxProvider>
          <ConfirmBoxProvider>{children}</ConfirmBoxProvider>
        </MsgBoxProvider>
      </ThemeProvider>
    </AppRouterCacheProvider>
  );
}
