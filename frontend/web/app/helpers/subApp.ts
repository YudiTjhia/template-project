export type SubAppType = 'admin' | 'customer';

export function resolveSubAppType(value: string | undefined): SubAppType {
  return value === 'customer' ? 'customer' : 'admin';
}

const subAppPaths: Record<SubAppType, { prefix: string; login: string; register: string; home: string }> = {
  admin: { prefix: '/admin', login: '/admin/login', register: '/admin/account-register', home: '/admin' },
  customer: { prefix: '/customer', login: '/customer/login', register: '/customer/account-register', home: '/customer' },
};

export function getSubAppPaths(subAppType: SubAppType) {
  return subAppPaths[subAppType];
}

/** Add the active sub-app prefix to an app-relative path. */
export function getSubAppUrl(subAppType: SubAppType, path = '/') {
  const { prefix } = getSubAppPaths(subAppType);
  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  if (normalizedPath === '/') return prefix;
  if (normalizedPath === prefix || normalizedPath.startsWith(`${prefix}/`)) return normalizedPath;
  return `${prefix}${normalizedPath}`;
}
