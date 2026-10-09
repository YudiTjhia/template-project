import type { LoginValues } from '@/app/components/login/LoginForm';
import type { SubAppType } from '@/app/helpers/subApp';
import { SubAppApi, SubAppApiError } from './SubAppApi';

export type LoginRequest = LoginValues;

export type LoginOTPRequest = {
  identifier: string;
};

export type VerifyLoginOTPRequest = LoginOTPRequest & {
  otp: string;
};

export type LoginResponse<T = unknown> = {
  data: T;
  message?: string;
};

export type LoginServiceOptions = {
  baseUrl?: string;
  loginEndpoint?: string;
  requestOTPEndpoint?: string;
  verifyOTPEndpoint?: string;
  logoutEndpoint?: string;
  validateSessionEndpoint?: string;
  fetcher?: typeof fetch;
};

export class LoginApiError extends Error {
  constructor(message: string, public readonly status: number, public readonly payload: unknown) {
    super(message);
    this.name = 'LoginApiError';
  }
}

const defaultEndpoints = {
  login: '/api/login',
  requestOTP: '/api/login/otp/request',
  verifyOTP: '/api/login/otp/verify',
  logout: '/api/login/logout',
  validateSession: '/api/login/validate-session',
};

const sessionTokenKey = 'auth_token';

/** Authentication API for the active sub-app, called through its same-origin Next.js routes. */
export class LoginService {
  private readonly api: SubAppApi;
  private readonly endpoints: typeof defaultEndpoints;

  constructor(options: LoginServiceOptions = {}) {
    this.api = new SubAppApi({ baseUrl: options.baseUrl ?? '', fetcher: options.fetcher });
    this.endpoints = {
      login: options.loginEndpoint ?? defaultEndpoints.login,
      requestOTP: options.requestOTPEndpoint ?? defaultEndpoints.requestOTP,
      verifyOTP: options.verifyOTPEndpoint ?? defaultEndpoints.verifyOTP,
      logout: options.logoutEndpoint ?? defaultEndpoints.logout,
      validateSession: options.validateSessionEndpoint ?? defaultEndpoints.validateSession,
    };
  }

  login<T = unknown>(credentials: LoginRequest): Promise<LoginResponse<T>> {
    return this.post<T>(this.endpoints.login, credentials, 'Login gagal.')
      .then((response) => this.rememberSessionToken(response));
  }

  requestLoginOTP<T = unknown>(identifier: string): Promise<LoginResponse<T>> {
    return this.post<T>(this.endpoints.requestOTP, { identifier }, 'Permintaan OTP login gagal.');
  }

  verifyLoginOTP<T = unknown>(identifier: string, otp: string): Promise<LoginResponse<T>> {
    return this.post<T>(this.endpoints.verifyOTP, { identifier, otp }, 'Verifikasi OTP login gagal.')
      .then((response) => this.rememberSessionToken(response));
  }

  /** Alias for verifying an OTP and completing OTP login. */
  loginOTP<T = unknown>(identifier: string, otp: string): Promise<LoginResponse<T>> {
    return this.verifyLoginOTP<T>(identifier, otp);
  }

  logout<T = unknown>(token: string): Promise<LoginResponse<T>> {
    return this.post<T>(this.endpoints.logout, { token }, 'Logout gagal.');
  }

  async logoutCurrentSession(): Promise<void> {
    if (typeof window === 'undefined') return;

    const token = this.getSessionToken();
    try {
      if (token) await this.logout(token);
    } finally {
      window.sessionStorage.removeItem(sessionTokenKey);
    }
  }

  getSessionToken(): string | null {
    if (typeof window === 'undefined') return null;
    return window.sessionStorage.getItem(sessionTokenKey);
  }

  validateSession<T = unknown>(token: string, subAppType: SubAppType): Promise<LoginResponse<T>> {
    return this.post<T>(this.endpoints.validateSession, { token, sub_app_type: subAppType }, 'Validasi sesi gagal.')
      .then((response) => this.rememberSessionToken(response));
  }

  clearSessionToken(): void {
    if (typeof window !== 'undefined') window.sessionStorage.removeItem(sessionTokenKey);
  }

  private async post<T>(endpoint: string, body: unknown, fallbackMessage: string): Promise<LoginResponse<T>> {
    let payload: unknown;
    try {
      payload = await this.api.post<unknown>(endpoint, body, fallbackMessage);
    } catch (error) {
      if (error instanceof SubAppApiError) {
        throw new LoginApiError(error.message, error.status, error.payload);
      }
      throw error;
    }

    if (payload && typeof payload === 'object' && 'data' in payload) {
      return payload as LoginResponse<T>;
    }
    return { data: payload as T };
  }

  private rememberSessionToken<T>(response: LoginResponse<T>): LoginResponse<T> {
    if (typeof window !== 'undefined') {
      const token = this.findToken(response.data);
      if (token) window.sessionStorage.setItem(sessionTokenKey, token);
    }
    return response;
  }

  private findToken(value: unknown): string | undefined {
    if (!value || typeof value !== 'object') return undefined;
    const record = value as Record<string, unknown>;
    const token = record.token ?? record.Token;
    if (typeof token === 'string' && token.trim()) return token;
    return this.findToken(record.data);
  }
}

const loginService = new LoginService();
export default loginService;
