'use client';

import { use, useMemo } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { RoleComponent } from '@/app/components/role';
import { useMsgBox } from '@/app/components/msgbox';

export default function RoleDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const roleId = useMemo(() => Number(id), [id]);
  const router = useRouter();
  const searchParams = useSearchParams();
  const { showMessage } = useMsgBox();
  const mode = searchParams.get('mode') === 'view' ? 'view' : 'edit';
  if (!Number.isInteger(roleId) || roleId <= 0) return null;
  return <RoleComponent mode={mode} roleId={roleId} onCancel={() => router.push('/roles')} onSaved={() => {
    showMessage({ message: 'Role berhasil diperbarui.', severity: 'success' });
    router.push('/roles');
  }} />;
}
