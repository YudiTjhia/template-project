import { proxyBackendRequest } from '@/app/services/backendProxy';

const backendPath = process.env.NEXT_PUBLIC_ACCOUNT_OTP_RESEND_PATH ?? '/auth/accounts/resend-otp';

export async function POST(request: Request) {
  return proxyBackendRequest(request, backendPath);
}
