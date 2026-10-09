import { SubAppApi } from './SubAppApi';
import loginService from './LoginService';

export type UserRecord = {
  id: number;
  username: string;
  user_email: string;
  mobilephone: string;
  user_type: string;
  login_type: string;
  user_roles: string[];
  is_active: boolean;
};

export type UserPayload = Omit<UserRecord, 'id'> & { user_password?: string };
type DataResponse<T> = { data?: T };
type IdResponse = { id?: number };

export type UserServiceOptions = {
  baseUrl?: string;
  fetcher?: typeof fetch;
  tokenProvider?: () => string | null;
};

export class UserService {
  private readonly api: SubAppApi;
  private readonly tokenProvider: () => string | null;

  constructor(options: UserServiceOptions = {}) {
    this.api = new SubAppApi({ baseUrl: options.baseUrl, fetcher: options.fetcher });
    this.tokenProvider = options.tokenProvider ?? (() => loginService.getSessionToken());
  }

  async list(search = '') {
    const result = await this.api.get<DataResponse<UserRecord[]>>('/api/users', 'Gagal mengambil daftar user.', {
      token: this.token(), query: { search },
    });
    return result.data ?? [];
  }

  async get(id: number) {
    const result = await this.api.get<DataResponse<UserRecord>>(`/api/users/${id}`, 'Gagal mengambil detail user.', {
      token: this.token(),
    });
    if (!result.data) throw new Error('Data user tidak ditemukan.');
    return result.data;
  }

  async create(payload: UserPayload) {
    return this.api.requestWithOptions<IdResponse>('/api/users', 'POST', payload, 'Gagal membuat user.', { token: this.token() });
  }

  async update(id: number, payload: UserPayload) {
    await this.api.requestWithOptions(`/api/users/${id}`, 'PUT', payload, 'Gagal memperbarui user.', { token: this.token() });
  }

  async remove(id: number) {
    await this.api.requestWithOptions(`/api/users/${id}`, 'DELETE', undefined, 'Gagal menghapus user.', { token: this.token() });
  }

  private token() {
    const token = this.tokenProvider();
    if (!token) throw new Error('Sesi login tidak ditemukan. Silakan login kembali.');
    return token;
  }
}

const userService = new UserService();
export default userService;
