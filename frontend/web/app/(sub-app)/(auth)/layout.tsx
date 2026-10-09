import type { ReactNode } from 'react';
import AuthedLayout from '@/app/components/layout/AuthedLayout';
import { resolveSubAppType } from '@/app/helpers/subApp';

export default function AuthLayout({ children }: { children: ReactNode }) {
  const subAppType = resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE);
  const appName = process.env.APP_NAME?.trim() || 'Application';
  return <AuthedLayout appName={appName} subAppType={subAppType}>{children}</AuthedLayout>;
}
