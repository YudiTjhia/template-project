'use client';

import { Box, CircularProgress, Stack, Typography } from '@mui/material';

export default function InitialSplashScreen() {
  return (
    <Box
      role="status"
      aria-label="Memuat aplikasi"
      sx={{
        position: 'fixed',
        inset: 0,
        zIndex: 2000,
        display: 'grid',
        placeItems: 'center',
        color: 'primary.contrastText',
        bgcolor: 'primary.main',
      }}
    >
      <Stack alignItems="center" spacing={2}>
        <Box
          aria-hidden="true"
          sx={{
            width: 64,
            height: 64,
            display: 'grid',
            placeItems: 'center',
            borderRadius: 3,
            bgcolor: 'common.white',
            color: 'primary.main',
            fontSize: 30,
            fontWeight: 800,
            boxShadow: '0 8px 24px rgba(0, 0, 0, 0.16)',
          }}
        >
          T
        </Box>
        <Typography variant="h6" fontWeight={700}>
          Template Project
        </Typography>
        <CircularProgress size={24} thickness={5} sx={{ color: 'common.white' }} />
      </Stack>
    </Box>
  );
}
