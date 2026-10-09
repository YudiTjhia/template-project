'use client';

import type { CSSProperties, ReactNode } from 'react';
import { Box } from '@mui/material';
import type { SxProps, Theme } from '@mui/material/styles';

export type PageSpacing = number | string | { xs?: number | string; sm?: number | string; md?: number | string; lg?: number | string; xl?: number | string };

export type PageContainerProps = {
  children: ReactNode;
  padding?: PageSpacing;
  margin?: PageSpacing;
  maxWidth?: number | string;
  fullWidth?: boolean;
  component?: 'main' | 'div' | 'section' | 'article';
  sx?: SxProps<Theme>;
  id?: string;
  className?: string;
  role?: string;
  style?: CSSProperties;
};

export default function PageContainer({
  children,
  padding = { xs: 2, md: 4 },
  margin = '0 auto',
  maxWidth = 1200,
  fullWidth = true,
  component = 'main',
  sx,
  ...otherProps
}: PageContainerProps) {
  return (
    <Box
      component={component}
      sx={{
        width: fullWidth ? '100%' : undefined,
        maxWidth,
        p: padding,
        m: margin,
        ...sx,
      }}
      {...otherProps}
    >
      {children}
    </Box>
  );
}
