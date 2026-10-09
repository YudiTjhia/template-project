import type { NextConfig } from 'next';

const subAppType = process.env.SUB_APP_TYPE ?? 'admin';
if (subAppType !== 'admin' && subAppType !== 'customer') {
  throw new Error(`SUB_APP_TYPE must be "admin" or "customer"; received "${subAppType}"`);
}

const selectedApiUrl = subAppType === 'admin'
  ? process.env.ADMIN_API_BASE_URL
  : process.env.CUSTOMER_API_BASE_URL;

const nextConfig: NextConfig = {
  output: 'standalone',
  env: {
    NEXT_PUBLIC_SUB_APP_TYPE: subAppType,
    NEXT_PUBLIC_API_BASE_URL: selectedApiUrl ?? process.env.NEXT_PUBLIC_API_BASE_URL ?? '',
  },
};

export default nextConfig;
