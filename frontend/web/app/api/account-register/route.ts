import { proxyBackendRequest } from '@/app/services/backendProxy';

const backendPath = process.env.NEXT_PUBLIC_ACCOUNT_REGISTER_PATH ?? '/auth/accounts/register';

export async function POST(request: Request) {
  return proxyBackendRequest(request, backendPath);
}
