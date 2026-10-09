import { proxyBackendRequest } from '@/app/services/backendProxy';

const backendPath = process.env.NEXT_PUBLIC_ACCOUNT_OTP_VERIFY_PATH ?? '/auth/accounts/verify';

export async function POST(request: Request) {
  return proxyBackendRequest(request, backendPath);
}
