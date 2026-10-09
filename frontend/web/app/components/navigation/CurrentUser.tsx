'use client';

import { useEffect, useState } from 'react';
import { Avatar, Box, Stack, Typography } from '@mui/material';
import loginService, { type CurrentSessionUser } from '@/app/services/LoginService';

export type CurrentUserProps = {
  user?: CurrentSessionUser | null;
  compact?: boolean;
};

export default function CurrentUser({ user: suppliedUser, compact = false }: CurrentUserProps) {
  const [sessionUser, setSessionUser] = useState<CurrentSessionUser | null>(suppliedUser ?? null);

  useEffect(() => {
    setSessionUser(suppliedUser === undefined ? loginService.getCurrentUser() : suppliedUser);
  }, [suppliedUser]);

  const username = sessionUser?.username || sessionUser?.email || 'Pengguna';
  const email = sessionUser?.email || 'Email tidak tersedia';
  const initials = username.trim().split(/\s+/).slice(0, 2).map((part) => part[0]).join('').toLocaleUpperCase('id');

  return (
    <Stack direction="row" spacing={1.25} alignItems="center" sx={{ px: 2, py: compact ? 1 : 1.5, minWidth: 0 }}>
      <Avatar sx={{ width: compact ? 36 : 42, height: compact ? 36 : 42, bgcolor: 'primary.main', fontSize: 14, fontWeight: 700 }}>
        {initials || 'U'}
      </Avatar>
      <Box sx={{ minWidth: 0, flex: 1 }}>
        <Typography variant="body2" fontWeight={700} noWrap title={username}>
          {username}
        </Typography>
        <Typography variant="caption" color="text.secondary" noWrap title={email}>
          {email}
        </Typography>
      </Box>
    </Stack>
  );
}
