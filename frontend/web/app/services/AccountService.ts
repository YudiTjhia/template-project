import { SubAppApi, SubAppApiError } from './SubAppApi';

/** Registration payload accepted by the backend account endpoint. */
export type AccountRegisterRequest = {
  account_name: string;
  account_email: string;
  address: string;
  mobilephone: string;
  password: string;
  confirm_password: string;
  pin: string;
  confirm_pin: string;
  terms_accepted: boolean;
};

export type AccountResponse<T = unknown> = {
  data: T;
  message?: string;
};

export type AccountServiceOptions = {
  baseUrl?: string;
  registerEndpoint?: string;
  verifyOTPEndpoint?: string;
  resendOTPEndpoint?: string;
  fetcher?: typeof fetch;
};

export class AccountApiError extends Error {
  constructor(message: string, public readonly status: number, public readonly payload: unknown) {
    super(message);
    this.name = 'AccountApiError';
  }
}

const configuredRegisterEndpoint = '/api/account-register';
const configuredVerifyOTPEndpoint = '/api/account-register/verify-otp';
const configuredResendOTPEndpoint = '/api/account-register/resend-otp';

/** Account registration and registration OTP API for the active sub-app. */
export class AccountService {
  private readonly api: SubAppApi;
  private readonly registerEndpoint: string;
  private readonly verifyOTPEndpoint: string;
  private readonly resendOTPEndpoint: string;

  constructor(options: AccountServiceOptions = {}) {
    this.api = new SubAppApi({ baseUrl: options.baseUrl ?? '', fetcher: options.fetcher });
    this.registerEndpoint = options.registerEndpoint ?? configuredRegisterEndpoint;
    this.verifyOTPEndpoint = options.verifyOTPEndpoint ?? configuredVerifyOTPEndpoint;
    this.resendOTPEndpoint = options.resendOTPEndpoint ?? configuredResendOTPEndpoint;
  }

  async register<T = unknown>(values: AccountRegisterRequest): Promise<AccountResponse<T>> {
    return this.post<T>(this.registerEndpoint, values, 'Pendaftaran akun gagal.');
  }

  async verifikasiOTP<T = unknown>(otp: string): Promise<AccountResponse<T>> {
    return this.post<T>(this.verifyOTPEndpoint, { otp }, 'Verifikasi OTP gagal.');
  }

  /** English alias for callers that prefer an English method name. */
  verifyOTP<T = unknown>(otp: string) {
    return this.verifikasiOTP<T>(otp);
  }

  async resendOTP<T = unknown>(accountId: number): Promise<AccountResponse<T>> {
    return this.post<T>(this.resendOTPEndpoint, { account_id: accountId }, 'Pengiriman ulang OTP gagal.');
  }

  private async post<T>(endpoint: string, body: unknown, fallbackMessage: string): Promise<AccountResponse<T>> {
    let payload: unknown;
    try {
      payload = await this.api.post<unknown>(endpoint, body, fallbackMessage);
    } catch (error) {
      if (error instanceof SubAppApiError) {
        throw new AccountApiError(error.message, error.status, error.payload);
      }
      throw error;
    }

    if (payload && typeof payload === 'object' && 'data' in payload) {
      return payload as AccountResponse<T>;
    }
    return { data: payload as T };
  }
}

const accountService = new AccountService();
export default accountService;
