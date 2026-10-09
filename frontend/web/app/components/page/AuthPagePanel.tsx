'use client';

import type { ReactNode } from 'react';
import { Paper } from '@mui/material';

export default function AuthPagePanel({ children }: { children: ReactNode }) {
  return (
    <Paper
      elevation={0}
      sx={{
        width: '100%',
        maxWidth: 520,
        p: { xs: 2, sm: 3 },
        bgcolor: 'common.white',
        borderRadius: 3,
        // boxShadow: '0 12px 32px rgba(8, 47, 96, 0.2)',
      }}
    >
      {children}
    </Paper>
  );
}
