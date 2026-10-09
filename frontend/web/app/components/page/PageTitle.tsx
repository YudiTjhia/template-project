'use client';

import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import { Box, IconButton, Stack, Typography } from '@mui/material';

export type PageTitleProps = {
  title: string;
  subtitle?: string;
  showBackButton?: boolean;
  onBack?: () => void;
  backButtonLabel?: string;
  /** Optional content aligned to the right of the title and subtitle. */
  actions?: React.ReactNode;
};

export default function PageTitle({
  title,
  subtitle,
  showBackButton = false,
  onBack,
  backButtonLabel = 'Kembali',
  actions,
}: PageTitleProps) {
  return (
    <Stack
      direction={{ xs: 'column', sm: 'row' }}
      justifyContent="space-between"
      alignItems={{ xs: 'stretch', sm: 'center' }}
      spacing={2}
      sx={{ mb: 3 }}
    >
      <Stack direction="row" alignItems="center" spacing={1.5}>
        {showBackButton && (
          <IconButton aria-label={backButtonLabel} onClick={onBack} edge="start">
            <ArrowBackIcon />
          </IconButton>
        )}
        <Box>
          <Typography variant="h4" component="h1" fontWeight={700}>
            {title}
          </Typography>
          {subtitle && (
            <Typography color="text.secondary" sx={{ mt: 0.5 }}>
              {subtitle}
            </Typography>
          )}
        </Box>
      </Stack>
      {actions}
    </Stack>
  );
}
