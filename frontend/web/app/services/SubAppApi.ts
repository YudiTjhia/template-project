/** Shared HTTP client for requests made by the active sub-app. */
export type SubAppApiOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

export class SubAppApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly payload: unknown,
  ) {
    super(message);
    this.name = 'SubAppApiError';
  }
}

const subAppApiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? '';

export function buildApiUrl(baseUrl: string, endpoint: string) {
  if (/^https?:\/\//i.test(endpoint)) return endpoint;
  if (!baseUrl) return endpoint;
  return `${baseUrl.replace(/\/$/, '')}/${endpoint.replace(/^\//, '')}`;
}

async function parseResponse(response: Response): Promise<unknown> {
  const contentType = response.headers.get('content-type') ?? '';
  if (contentType.includes('application/json')) {
    try {
      return await response.json();
    } catch {
      return undefined;
    }
  }
  return response.text();
}

function responseMessage(payload: unknown, fallback: string) {
  if (typeof payload === 'string' && payload.trim()) return payload;
  if (payload && typeof payload === 'object' && 'message' in payload) {
    const message = (payload as { message?: unknown }).message;
    if (typeof message === 'string' && message.trim()) return message;
  }
  return fallback;
}

export class SubAppApi {
  private readonly baseUrl: string;
  private readonly fetcher: typeof fetch;

  constructor(options: SubAppApiOptions = {}) {
    this.baseUrl = options.baseUrl ?? subAppApiBaseUrl;
    this.fetcher = options.fetcher ?? fetch;
  }

  async post<T>(endpoint: string, body: unknown, fallbackMessage: string): Promise<T> {
    const response = await this.fetcher(buildApiUrl(this.baseUrl, endpoint), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
      body: JSON.stringify(body),
    });
    const payload = await parseResponse(response);

    if (!response.ok) {
      throw new SubAppApiError(responseMessage(payload, fallbackMessage), response.status, payload);
    }
    return payload as T;
  }
}

const subAppApi = new SubAppApi();
export default subAppApi;
