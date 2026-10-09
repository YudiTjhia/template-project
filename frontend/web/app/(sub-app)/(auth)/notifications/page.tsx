'use client';

import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Box,
  Button,
  Chip,
  CircularProgress,
  Divider,
  List,
  ListItem,
  ListItemText,
  Paper,
  Stack,
  Typography,
} from '@mui/material';
import MarkEmailReadOutlinedIcon from '@mui/icons-material/MarkEmailReadOutlined';
import RefreshOutlinedIcon from '@mui/icons-material/RefreshOutlined';
import PageContainer from '@/app/components/page/PageContainer';
import PageTitle from '@/app/components/page/PageTitle';
import SearchBox from '@/app/components/list/SearchBox';
import { useMsgBox } from '@/app/components/msgbox';
import notificationService, { type NotificationItem } from '@/app/services/NotificationService';

function formatDate(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium', timeStyle: 'short' }).format(date);
}

export default function NotificationsPage() {
  const { showMessage } = useMsgBox();
  const [notifications, setNotifications] = useState<NotificationItem[]>([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [activeId, setActiveId] = useState<number>();
  const [error, setError] = useState<string>();

  const loadNotifications = useCallback(async (query: string) => {
    setLoading(true);
    setError(undefined);
    try {
      const [items, unread] = await Promise.all([
        notificationService.getNotifications({ search: query, limit: 100 }),
        notificationService.getUnreadCount(),
      ]);
      setNotifications(items);
      setUnreadCount(unread);
    } catch (caughtError) {
      setError(caughtError instanceof Error ? caughtError.message : 'Gagal memuat notifikasi.');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadNotifications('');
  }, [loadNotifications]);

  useEffect(() => {
    let unsubscribe: (() => void) | undefined;
    void notificationService.listenForForegroundMessages((payload) => {
      const title = payload.notification?.title ?? 'Notifikasi baru';
      const message = payload.notification?.body ?? 'Anda menerima notifikasi baru.';
      showMessage({ title, message, severity: 'info' });
      void loadNotifications(search);
    }).then((stop) => {
      unsubscribe = stop;
    }).catch(() => undefined);
    return () => unsubscribe?.();
  }, [loadNotifications, search, showMessage]);

  const handleMarkRead = async (item: NotificationItem) => {
    setActiveId(item.id);
    try {
      await notificationService.markAsRead(item.id);
      setNotifications((current) => current.map((notification) =>
        notification.id === item.id ? { ...notification, isRead: true } : notification,
      ));
      setUnreadCount((current) => Math.max(0, current - 1));
    } catch (caughtError) {
      showMessage({
        title: 'Gagal memperbarui notifikasi',
        message: caughtError instanceof Error ? caughtError.message : 'Coba lagi.',
        severity: 'error',
      });
    } finally {
      setActiveId(undefined);
    }
  };

  return (
    <PageContainer maxWidth={960}>
      <PageTitle
        title="Notifikasi"
        subtitle="Lihat pembaruan dan aktivitas terbaru pada akun Anda."
        actions={(
          <Stack direction="row" alignItems="center" spacing={1}>
            <Chip label={`${unreadCount} belum dibaca`} color={unreadCount ? 'primary' : 'default'} variant="outlined" />
            <Button onClick={() => void loadNotifications(search)} startIcon={<RefreshOutlinedIcon />} disabled={loading}>
              Muat ulang
            </Button>
          </Stack>
        )}
      />

      <Box sx={{ mb: 2, maxWidth: 420 }}>
        <SearchBox
          value={search}
          placeholder="Cari notifikasi"
          onChange={(value) => {
            setSearch(value);
            void loadNotifications(value);
          }}
        />
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}

      <Paper variant="outlined" sx={{ borderRadius: 2, overflow: 'hidden' }}>
        {loading ? (
          <Stack alignItems="center" justifyContent="center" spacing={1.5} sx={{ py: 7 }}>
            <CircularProgress size={28} />
            <Typography color="text.secondary">Memuat notifikasi…</Typography>
          </Stack>
        ) : notifications.length === 0 ? (
          <Typography color="text.secondary" textAlign="center" sx={{ py: 7 }}>
            Belum ada notifikasi.
          </Typography>
        ) : (
          <List disablePadding>
            {notifications.map((item, index) => (
              <Box key={item.id}>
                {index > 0 && <Divider component="li" />}
                <ListItem
                  alignItems="flex-start"
                  secondaryAction={!item.isRead && (
                    <Button
                      size="small"
                      startIcon={activeId === item.id ? <CircularProgress size={14} /> : <MarkEmailReadOutlinedIcon />}
                      onClick={() => void handleMarkRead(item)}
                      disabled={activeId !== undefined}
                    >
                      Tandai dibaca
                    </Button>
                  )}
                  sx={{ bgcolor: item.isRead ? 'background.paper' : 'action.hover', py: 2 }}
                >
                  <ListItemText
                    primary={(
                      <Stack direction="row" alignItems="center" spacing={1} sx={{ pr: 2 }}>
                        <Typography component="span" fontWeight={item.isRead ? 500 : 700}>{item.title}</Typography>
                        {!item.isRead && <Chip size="small" color="primary" label="Baru" />}
                      </Stack>
                    )}
                    secondary={(
                      <>
                        <Typography component="span" variant="body2" color="text.primary" display="block" sx={{ mt: 0.5 }}>
                          {item.message}
                        </Typography>
                        <Typography component="span" variant="caption" color="text.secondary" display="block" sx={{ mt: 0.75 }}>
                          {item.entityNumber ? `${item.entityNumber} · ` : ''}{formatDate(item.createdAt)}
                        </Typography>
                      </>
                    )}
                  />
                </ListItem>
              </Box>
            ))}
          </List>
        )}
      </Paper>
    </PageContainer>
  );
}
