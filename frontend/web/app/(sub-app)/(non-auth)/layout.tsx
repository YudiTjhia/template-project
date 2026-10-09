import type { ReactNode } from 'react';
import PublicLayout from '@/app/components/layout/PublicLayout';
import { resolveSubAppType } from '@/app/helpers/subApp';

export default function NonAuthLayout({ children }: { children: ReactNode }) {
  const subAppType = resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE);
  return (
    <PublicLayout subAppType={subAppType} showHeader={false}>
      {children}
    </PublicLayout>
  );
}
