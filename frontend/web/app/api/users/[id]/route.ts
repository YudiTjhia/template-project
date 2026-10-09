import { proxyBackendRequest } from '@/app/services/backendProxy';

export async function GET(request: Request, context: { params: Promise<{ id: string }> }) {
  const params = await context.params;
  return proxyBackendRequest(request, '/users/[id]'.replace('[id]', encodeURIComponent(params.id)));
}

export async function PUT(request: Request, context: { params: Promise<{ id: string }> }) {
  const params = await context.params;
  return proxyBackendRequest(request, '/users/[id]'.replace('[id]', encodeURIComponent(params.id)));
}

export async function DELETE(request: Request, context: { params: Promise<{ id: string }> }) {
  const params = await context.params;
  return proxyBackendRequest(request, '/users/[id]'.replace('[id]', encodeURIComponent(params.id)));
}

