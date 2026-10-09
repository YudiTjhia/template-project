'use client';

import CssBaseline from '@mui/material/CssBaseline';
import { ThemeProvider } from '@mui/material/styles';
import type { ReactNode } from 'react';
import { getSubAppTheme, type SubAppType } from '../../themes/theme';

export default function SubAppThemeProvider({
  subApp,
  children,
}: {
  subApp: SubAppType;
  children: ReactNode;
}) {
  return (
    <ThemeProvider theme={getSubAppTheme(subApp)}>
      <CssBaseline />
      {children}
    </ThemeProvider>
  );
}
