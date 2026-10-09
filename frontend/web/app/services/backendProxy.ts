import { NextResponse } from 'next/server';

/** Forward a same-origin Next API request to the backend selected for this build. */
export async function proxyBackendRequest(request: Request, backendPath: string) {
  const backendBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? '';
  if (!backendBaseUrl) {
    return NextResponse.json({ message: 'Backend API URL belum dikonfigurasi.' }, { status: 503 });
  }

  let body: string | undefined;
  let subAppType: string | undefined;
  if (!['GET', 'HEAD'].includes(request.method)) {
    const rawBody = await request.text();
    if (rawBody) {
      body = rawBody;
      if (request.headers.get('content-type')?.includes('application/json')) {
        try {
          const payload: unknown = JSON.parse(rawBody);
          if (payload && typeof payload === 'object' && 'sub_app_type' in payload) {
            const value = (payload as { sub_app_type?: unknown }).sub_app_type;
            if (value === 'admin' || value === 'customer') subAppType = value;
          }
        } catch {
          return NextResponse.json({ message: 'Request body harus berupa JSON.' }, { status: 400 });
        }
      }
    }
  }

  const incomingUrl = new URL(request.url);
  const queryString = incomingUrl.search;
  const url = `${backendBaseUrl.replace(/\/$/, '')}/${backendPath.replace(/^\//, '')}${queryString}`;
  try {
    const backendResponse = await fetch(url, {
      method: request.method,
      headers: {
        Accept: 'application/json',
        ...(body !== undefined ? { 'Content-Type': request.headers.get('content-type') ?? 'application/json' } : {}),
        ...(request.headers.get('authorization') ? { Authorization: request.headers.get('authorization')! } : {}),
        'X-Sub-App-Type': subAppType ?? process.env.NEXT_PUBLIC_SUB_APP_TYPE ?? 'admin',
      },
      ...(body !== undefined ? { body } : {}),
      cache: 'no-store',
    });
    const responseBody = await backendResponse.text();
    return new NextResponse(responseBody, {
      status: backendResponse.status,
      headers: {
        'Content-Type': backendResponse.headers.get('content-type') ?? 'application/json; charset=utf-8',
      },
    });
  } catch {
    return NextResponse.json({ message: 'Tidak dapat terhubung ke backend.' }, { status: 502 });
  }
}
