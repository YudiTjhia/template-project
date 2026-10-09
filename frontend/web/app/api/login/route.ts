import { proxyBackendRequest } from '@/app/services/backendProxy';

const backendPath = process.env.NEXT_PUBLIC_LOGIN_PATH ?? '/auth/login';

export async function POST(request: Request) {
  return proxyBackendRequest(request, backendPath);
}
