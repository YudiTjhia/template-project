/** Shared HTTP client for requests made by the active sub-app. */
export type SubAppApiOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
};

export type ApiRequestOptions = {
  token?: string;
  query?: Record<string, string | number | boolean | undefined>;
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
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
  if (payload && typeof payload === 'object' && ('message' in payload || 'error' in payload)) {
    const record = payload as { message?: unknown; error?: unknown };
    const message = record.message ?? record.error;
    if (typeof message === 'string' && message.trim()) return message;
  }
  return fallback;
}

function getHeaders(token?: string): HeadersInit {
  return {
    Accept: 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export class SubAppApi {
  private readonly baseUrl: string;
  private readonly fetcher: typeof fetch;

  constructor(options: SubAppApiOptions = {}) {
    this.baseUrl = options.baseUrl ?? subAppApiBaseUrl;
    this.fetcher = options.fetcher ?? fetch;
  }

  async post<T>(endpoint: string, body: unknown, fallbackMessage: string): Promise<T> {
    return this.request<T>(endpoint, 'POST', body, fallbackMessage);
  }

  async get<T>(endpoint: string, fallbackMessage: string, options: ApiRequestOptions = {}): Promise<T> {
    const query = new URLSearchParams();
    Object.entries(options.query ?? {}).forEach(([key, value]) => {
      if (value !== undefined) query.set(key, String(value));
    });
    const queryString = query.toString();
    const endpointWithQuery = queryString
      ? `${endpoint}${endpoint.includes('?') ? '&' : '?'}${queryString}`
      : endpoint;
    return this.request<T>(endpointWithQuery, 'GET', undefined, fallbackMessage, options.token);
  }

  async postWithOptions<T>(
    endpoint: string,
    body: unknown,
    fallbackMessage: string,
    options: ApiRequestOptions = {},
  ): Promise<T> {
    return this.request<T>(endpoint, 'POST', body, fallbackMessage, options.token);
  }

  async requestWithOptions<T>(
    endpoint: string,
    method: NonNullable<ApiRequestOptions['method']>,
    body: unknown,
    fallbackMessage: string,
    options: ApiRequestOptions = {},
  ): Promise<T> {
    return this.request<T>(endpoint, method, body, fallbackMessage, options.token);
  }

  private async request<T>(
    endpoint: string,
    method: NonNullable<ApiRequestOptions['method']>,
    body: unknown,
    fallbackMessage: string,
    token?: string,
  ): Promise<T> {
    const response = await this.fetcher(buildApiUrl(this.baseUrl, endpoint), {
      method,
      headers: {
        ...getHeaders(token),
        ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}),
      },
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
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
