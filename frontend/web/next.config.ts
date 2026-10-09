import type { NextConfig } from 'next';

const subAppType = process.env.SUB_APP_TYPE ?? 'admin';
if (subAppType !== 'admin' && subAppType !== 'customer') {
  throw new Error(`SUB_APP_TYPE must be "admin" or "customer"; received "${subAppType}"`);
}

const nextConfig: NextConfig = {
  output: 'standalone',
  env: {
    NEXT_PUBLIC_SUB_APP_TYPE: subAppType,
  },
};

export default nextConfig;
