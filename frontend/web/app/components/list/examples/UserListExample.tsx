'use client';

import { Chip } from '@mui/material';
import ListComponent, { type ListColumn, type ListRow } from '../ListComponent';

type UserRow = ListRow & {
  name: string;
  email: string;
  status: 'active' | 'inactive';
  createdAt: string;
};

const rows: UserRow[] = [
  { id: 1, name: 'Andi Pratama', email: 'andi@example.com', status: 'active', createdAt: '2026-10-09' },
  { id: 2, name: 'Siti Rahmawati', email: 'siti@example.com', status: 'active', createdAt: '2026-10-08' },
  { id: 3, name: 'Budi Santoso', email: 'budi@example.com', status: 'inactive', createdAt: '2026-10-07' },
];

const columns: ListColumn<UserRow>[] = [
  { field: 'name', label: 'Nama' },
  { field: 'email', label: 'Email' },
  {
    field: 'status',
    label: 'Status',
    searchable: false,
    render: (value) => (
      <Chip
        size="small"
        label={value === 'active' ? 'Aktif' : 'Nonaktif'}
        color={value === 'active' ? 'success' : 'default'}
        variant="outlined"
      />
    ),
  },
  { field: 'createdAt', label: 'Tanggal dibuat' },
];

export default function UserListExample() {
  return (
    <ListComponent<UserRow>
      title="Daftar Pengguna"
      description="Kelola pengguna yang terdaftar."
      columns={columns}
      rows={rows}
      filters={[
        {
          field: 'status',
          label: 'Status',
          options: [
            { label: 'Aktif', value: 'active' },
            { label: 'Nonaktif', value: 'inactive' },
          ],
        },
      ]}
      sortableFields={['name', 'createdAt']}
      initialSortField="createdAt"
      initialSortDirection="desc"
      searchPlaceholder="Cari nama atau email"
      onAdd={() => window.alert('Buka form tambah pengguna')}
      onEdit={(user) => window.alert(`Edit ${user.name}`)}
      onDelete={(user) => window.alert(`Hapus ${user.name}`)}
    />
  );
}
