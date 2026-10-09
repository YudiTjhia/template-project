'use client';

import { use, useMemo } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { UserComponent } from '@/app/components/user';
import { useMsgBox } from '@/app/components/msgbox';

export default function UserDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const userId = useMemo(() => Number(id), [id]);
  const router = useRouter();
  const searchParams = useSearchParams();
  const { showMessage } = useMsgBox();
  const mode = searchParams.get('mode') === 'view' ? 'view' : 'edit';
  if (!Number.isInteger(userId) || userId <= 0) return null;
  return <UserComponent mode={mode} userId={userId} onCancel={() => router.push('/users')} onSaved={() => {
    showMessage({ message: 'User berhasil diperbarui.', severity: 'success' });
    router.push('/users');
  }} />;
}
