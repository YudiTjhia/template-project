'use client';

import { useRouter } from 'next/navigation';
import { RoleComponent } from '@/app/components/role';
import { useMsgBox } from '@/app/components/msgbox';

export default function NewRolePage() {
  const router = useRouter();
  const { showMessage } = useMsgBox();
  return <RoleComponent mode="create" onCancel={() => router.push('/roles')} onSaved={() => {
    showMessage({ message: 'Role berhasil dibuat.', severity: 'success' });
    router.push('/roles');
  }} />;
}
