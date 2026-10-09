import { proxyBackendRequest } from '@/app/services/backendProxy';

export async function POST(request: Request) {
  return proxyBackendRequest(request, '/auth/users/verify');
}
