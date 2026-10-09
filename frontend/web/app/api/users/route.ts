import { proxyBackendRequest } from '@/app/services/backendProxy';

export async function GET(request: Request) {
  return proxyBackendRequest(request, '/users');
}

export async function POST(request: Request) {
  return proxyBackendRequest(request, '/users');
}

