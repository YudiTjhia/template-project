import api from './SubAppApi';
import loginService from './LoginService';

export type RoleRecord = {
  id: number;
  role_name: string;
  actions: string[];
};

export type RoleActionRecord = {
  id: number;
  value: string;
  label: string;
  sort_order: number;
  active: boolean;
};

type DataResponse<T> = { data?: T };
type IdResponse = { id?: number };

export class RoleService {
  async list(search = '') {
    const result = await api.get<DataResponse<RoleRecord[]>>('/api/roles', 'Gagal mengambil daftar role.', {
      token: this.token(), query: { search },
    });
    return result.data ?? [];
  }

  async get(id: number) {
    const result = await api.get<DataResponse<RoleRecord>>(`/api/roles/${id}`, 'Gagal mengambil detail role.', {
      token: this.token(),
    });
    if (!result.data) throw new Error('Data role tidak ditemukan.');
    return result.data;
  }

  async listActions(search = '') {
    const result = await api.get<DataResponse<RoleActionRecord[]>>('/api/roles/actions', 'Gagal mengambil role action.', {
      token: this.token(), query: { search },
    });
    return result.data ?? [];
  }

  async create(payload: Omit<RoleRecord, 'id'>) {
    return api.requestWithOptions<IdResponse>('/api/roles', 'POST', payload, 'Gagal membuat role.', { token: this.token() });
  }

  async update(id: number, payload: Omit<RoleRecord, 'id'>) {
    await api.requestWithOptions(`/api/roles/${id}`, 'PUT', payload, 'Gagal memperbarui role.', { token: this.token() });
  }

  async remove(id: number) {
    await api.requestWithOptions(`/api/roles/${id}`, 'DELETE', undefined, 'Gagal menghapus role.', { token: this.token() });
  }

  private token() {
    const token = loginService.getSessionToken();
    if (!token) throw new Error('Sesi login tidak ditemukan. Silakan login kembali.');
    return token;
  }
}

const roleService = new RoleService();
export default roleService;
