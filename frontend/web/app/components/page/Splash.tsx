'use client';

import Link from 'next/link';
import ArrowForwardRoundedIcon from '@mui/icons-material/ArrowForwardRounded';
import { Box, Button, Container, Paper, Stack, Typography } from '@mui/material';
import { getSubAppUrl, type SubAppType } from '@/app/helpers/subApp';

export interface SplashProps {
  subAppType: SubAppType;
  title?: string;
  description?: string;
  showRegisterButton?: boolean;
  appName?: string;
}

export default function Splash({
  subAppType,
  title = 'Selamat datang',
  description,
  showRegisterButton = subAppType === 'admin',
  appName = 'Template Project',
}: SplashProps) {
  const isCustomer = subAppType === 'customer';

  return (
    <Box
      component="main"
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        py: { xs: 5, md: 8 },
        background:
          'radial-gradient(circle at 15% 15%, rgba(25, 118, 210, 0.13), transparent 34%), linear-gradient(145deg, #f7faff 0%, #eef3fa 100%)',
      }}
    >
      <Container maxWidth="md">
        <Paper
          elevation={0}
          sx={{
            overflow: 'hidden',
            border: '1px solid',
            borderColor: 'divider',
            borderRadius: { xs: 3, md: 5 },
            backgroundColor: 'rgba(255, 255, 255, 0.94)',
          }}
        >
          <Stack spacing={3} alignItems="flex-start" sx={{ p: { xs: 4, sm: 6, md: 8 } }}>
            <Box
              aria-hidden="true"
              sx={{
                width: 52,
                height: 52,
                borderRadius: 2.5,
                display: 'grid',
                placeItems: 'center',
                color: 'primary.contrastText',
                fontWeight: 800,
                fontSize: 22,
                background: (theme) => `linear-gradient(135deg, ${theme.palette.primary.dark}, ${theme.palette.primary.main})`,
              }}
            >
              {appName.charAt(0).toUpperCase()}
            </Box>

            <Box>
              <Typography variant="overline" color="primary.main" fontWeight={700}>
                {isCustomer ? 'CUSTOMER PORTAL' : 'ADMIN WORKSPACE'}
              </Typography>
              <Typography
                component="h1"
                variant="h2"
                fontWeight={700}
                sx={{ mt: 0.5, fontSize: { xs: '2.25rem', sm: '3.25rem' }, letterSpacing: '-0.04em' }}
              >
                {title}
              </Typography>
              <Typography color="text.secondary" sx={{ mt: 1.5, maxWidth: 560, fontSize: { xs: 16, sm: 18 } }}>
                {description ??
                  (isCustomer
                    ? 'Masuk ke portal pelanggan untuk melanjutkan.'
                    : 'Kelola pengguna, peran, dan aktivitas aplikasi dari satu tempat.')}
              </Typography>
            </Box>

            <Stack
              direction={{ xs: 'column', sm: 'row' }}
              spacing={1.5}
              sx={{ pt: 1, width: { xs: '100%', sm: 'auto' } }}
            >
              <Button
                component={Link}
                href={getSubAppUrl(subAppType, '/login')}
                variant="contained"
                size="large"
                endIcon={<ArrowForwardRoundedIcon />}
                sx={{ px: 3, py: 1.25, borderRadius: 2, textTransform: 'none', fontWeight: 700 }}
              >
                Masuk
              </Button>
              {showRegisterButton && (
                <Button
                  component={Link}
                  href={getSubAppUrl(subAppType, '/account-register')}
                  variant="outlined"
                  size="large"
                  sx={{ px: 3, py: 1.25, borderRadius: 2, textTransform: 'none', fontWeight: 600 }}
                >
                  Buat akun
                </Button>
              )}
            </Stack>
          </Stack>
          <Box
            sx={{
              height: 6,
              background: (theme) =>
                `linear-gradient(90deg, ${theme.palette.primary.dark}, ${theme.palette.primary.main}, ${theme.palette.primary.light})`,
            }}
          />
        </Paper>
        <Typography variant="body2" color="text.secondary" textAlign="center" sx={{ mt: 3 }}>
          {appName}
        </Typography>
      </Container>
    </Box>
  );
}
