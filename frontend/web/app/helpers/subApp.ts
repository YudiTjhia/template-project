export type SubAppType = 'admin' | 'customer';

export function resolveSubAppType(value: string | undefined): SubAppType {
  return value === 'customer' ? 'customer' : 'admin';
}

const subAppPaths: Record<SubAppType, { login: string; home: string }> = {
  admin: { login: '/login', home: '/' },
  customer: { login: '/login', home: '/' },
};

export function getSubAppPaths(subAppType: SubAppType) {
  return subAppPaths[subAppType];
}
