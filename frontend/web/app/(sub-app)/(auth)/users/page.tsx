'use client';

import { useCallback, useEffect, useState } from 'react';
import { Alert, CircularProgress, Stack } from '@mui/material';
import { useRouter } from 'next/navigation';
import ListComponent, { type ListColumn } from '@/app/components/list/ListComponent';
import { useConfirmBox } from '@/app/components/confirmbox/ConfirmBoxContext';
import { useMsgBox } from '@/app/components/msgbox';
import userService, { type UserRecord } from '@/app/services/UserService';

export default function UsersPage() {
  const router = useRouter();
  const { confirm } = useConfirmBox();
  const { showMessage } = useMsgBox();
  const [users, setUsers] = useState<UserRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try { setUsers(await userService.list()); }
    catch (caught) { setError(caught instanceof Error ? caught.message : 'Gagal memuat user.'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const remove = async (user: UserRecord) => {
    const accepted = await confirm({ title: 'Hapus user?', message: `User “${user.username}” akan dihapus.`, confirmColor: 'error', confirmLabel: 'Hapus user' });
    if (!accepted) return;
    try { await userService.remove(user.id); showMessage({ message: 'User berhasil dihapus.', severity: 'success' }); await load(); }
    catch (caught) { showMessage({ message: caught instanceof Error ? caught.message : 'Gagal menghapus user.', severity: 'error' }); }
  };

  const columns: ListColumn<UserRecord>[] = [
    { field: 'username', label: 'Username' },
    { field: 'user_email', label: 'Email' },
    { field: 'mobilephone', label: 'Telepon' },
    { field: 'user_type', label: 'Tipe' },
    { field: 'is_active', label: 'Status', render: (value) => value ? 'Aktif' : 'Nonaktif' },
  ];

  if (loading) return <Stack alignItems="center" sx={{ py: 8 }}><CircularProgress /></Stack>;
  return <>
    {error && <Alert severity="error" sx={{ mx: 3, mt: 2 }}>{error}</Alert>}
    <ListComponent
      title="Users"
      description="Kelola user dan hak aksesnya."
      columns={columns}
      rows={users}
      addLabel="Tambah user"
      onAdd={() => router.push('/users/new')}
      onView={(user) => router.push(`/users/${user.id}?mode=view`)}
      onEdit={(user) => router.push(`/users/${user.id}`)}
      onDelete={(user) => void remove(user)}
      getRowLabel={(user) => user.username}
      emptyMessage="Belum ada user."
    />
  </>;
}
