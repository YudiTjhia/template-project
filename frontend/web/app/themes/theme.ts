import type { Theme } from '@mui/material/styles';
import adminTheme from '@/app/(sub-app)/theme/theme';
import customerTheme from '@/app/(sub-app2)/theme/theme';

export type SubAppType = 'admin' | 'customer';

export const themes: Record<SubAppType, Theme> = {
  admin: adminTheme,
  customer: customerTheme,
};

export function getSubAppType(value: string | undefined): SubAppType {
  return value === 'customer' ? 'customer' : 'admin';
}

export function getSubAppTheme(value: string | undefined): Theme {
  return themes[getSubAppType(value)];
}

export default getSubAppTheme(process.env.NEXT_PUBLIC_SUB_APP_TYPE);
