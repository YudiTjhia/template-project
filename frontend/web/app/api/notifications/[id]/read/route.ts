import { proxyBackendRequest } from '@/app/services/backendProxy';

export async function POST(
  request: Request,
  context: { params: Promise<{ id: string }> },
) {
  const { id } = await context.params;
  if (!/^\d+$/.test(id)) {
    return Response.json({ message: 'ID notifikasi tidak valid.' }, { status: 400 });
  }
  return proxyBackendRequest(request, `/notifications/${id}/read`);
}
