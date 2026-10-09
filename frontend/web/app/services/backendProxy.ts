import { NextResponse } from 'next/server';

/** Forward a same-origin Next API request to the backend selected for this build. */
export async function proxyBackendRequest(request: Request, backendPath: string) {
  const backendBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? '';
  if (!backendBaseUrl) {
    return NextResponse.json({ message: 'Backend API URL belum dikonfigurasi.' }, { status: 503 });
  }

  let body: string;
  let subAppType: string | undefined;
  try {
    const payload: unknown = await request.json();
    body = JSON.stringify(payload);
    if (payload && typeof payload === 'object' && 'sub_app_type' in payload) {
      const value = (payload as { sub_app_type?: unknown }).sub_app_type;
      if (value === 'admin' || value === 'customer') subAppType = value;
    }
  } catch {
    return NextResponse.json({ message: 'Request body harus berupa JSON.' }, { status: 400 });
  }

  const url = `${backendBaseUrl.replace(/\/$/, '')}/${backendPath.replace(/^\//, '')}`;
  try {
    const backendResponse = await fetch(url, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Accept: 'application/json',
        'X-Sub-App-Type': subAppType ?? process.env.NEXT_PUBLIC_SUB_APP_TYPE ?? 'admin',
      },
      body,
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
