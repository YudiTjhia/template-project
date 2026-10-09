'use client';

import { useRouter } from 'next/navigation';
import { UserComponent } from '@/app/components/user';
import { useMsgBox } from '@/app/components/msgbox';

export default function NewUserPage() {
  const router = useRouter();
  const { showMessage } = useMsgBox();
  return <UserComponent mode="create" onCancel={() => router.push('/users')} onSaved={() => {
    showMessage({ message: 'User berhasil dibuat.', severity: 'success' });
    router.push('/users');
  }} />;
}
