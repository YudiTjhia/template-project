'use client';

import { useCallback, useEffect, useState } from 'react';
import { Alert, CircularProgress, Stack } from '@mui/material';
import { useRouter } from 'next/navigation';
import ListComponent, { type ListColumn } from '@/app/components/list/ListComponent';
import { useConfirmBox } from '@/app/components/confirmbox/ConfirmBoxContext';
import { useMsgBox } from '@/app/components/msgbox';
import roleService, { type RoleRecord } from '@/app/services/RoleService';

export default function RolesPage() {
  const router = useRouter();
  const { confirm } = useConfirmBox();
  const { showMessage } = useMsgBox();
  const [roles, setRoles] = useState<RoleRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>();

  const load = useCallback(async () => {
    setLoading(true);
    setError(undefined);
    try { setRoles(await roleService.list()); }
    catch (caught) { setError(caught instanceof Error ? caught.message : 'Gagal memuat role.'); }
    finally { setLoading(false); }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const remove = async (role: RoleRecord) => {
    const accepted = await confirm({ title: 'Hapus role?', message: `Role “${role.role_name}” akan dihapus.`, confirmColor: 'error', confirmLabel: 'Hapus role' });
    if (!accepted) return;
    try { await roleService.remove(role.id); showMessage({ message: 'Role berhasil dihapus.', severity: 'success' }); await load(); }
    catch (caught) { showMessage({ message: caught instanceof Error ? caught.message : 'Gagal menghapus role.', severity: 'error' }); }
  };

  const columns: ListColumn<RoleRecord>[] = [
    { field: 'role_name', label: 'Nama role' },
    { field: 'actions', label: 'Jumlah action', render: (value) => String(Array.isArray(value) ? value.length : 0) },
  ];

  if (loading) return <Stack alignItems="center" sx={{ py: 8 }}><CircularProgress /></Stack>;
  return <>
    {error && <Alert severity="error" sx={{ mx: 3, mt: 2 }}>{error}</Alert>}
    <ListComponent
      title="Roles"
      description="Kelola role dan action yang dapat dijalankan."
      columns={columns}
      rows={roles}
      addLabel="Tambah role"
      onAdd={() => router.push('/roles/new')}
      onView={(role) => router.push(`/roles/${role.id}?mode=view`)}
      onEdit={(role) => router.push(`/roles/${role.id}`)}
      onDelete={(role) => void remove(role)}
      getRowLabel={(role) => role.role_name}
      emptyMessage="Belum ada role."
    />
  </>;
}
