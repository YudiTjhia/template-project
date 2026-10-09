'use client';

import { useEffect, useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Checkbox,
  CircularProgress,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  FormControlLabel,
  Stack,
  Typography,
} from '@mui/material';
import SearchBox from '@/app/components/list/SearchBox';
import ListComponent, { type ListColumn } from '@/app/components/list/ListComponent';
import roleService, { type RoleActionRecord } from '@/app/services/RoleService';

export type RoleActionLookupDialogProps = {
  open: boolean;
  selectedValues: string[];
  onClose: () => void;
  onApply: (actions: RoleActionRecord[]) => void;
};

export default function RoleActionLookupDialog({ open, selectedValues, onClose, onApply }: RoleActionLookupDialogProps) {
  const [actions, setActions] = useState<RoleActionRecord[]>([]);
  const [selected, setSelected] = useState<string[]>(selectedValues);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>();

  useEffect(() => {
    if (!open) return;
    setSelected(selectedValues);
    setLoading(true);
    setError(undefined);
    roleService.listActions().then(setActions).catch((caught) => {
      setError(caught instanceof Error ? caught.message : 'Gagal memuat role action.');
    }).finally(() => setLoading(false));
  }, [open, selectedValues]);

  const visibleActions = useMemo(() => {
    const term = search.trim().toLocaleLowerCase('id');
    return actions.filter((action) => action.active && (!term || `${action.value} ${action.label}`.toLocaleLowerCase('id').includes(term)));
  }, [actions, search]);

  const columns: ListColumn<RoleActionRecord>[] = [
    { field: 'value', label: 'Kode' },
    { field: 'label', label: 'Nama' },
    {
      field: 'id', label: 'Pilih', searchable: false,
      render: (_value, row) => (
        <Checkbox
          checked={selected.includes(row.value)}
          onChange={(event) => setSelected((current) => event.target.checked
            ? [...new Set([...current, row.value])]
            : current.filter((value) => value !== row.value))}
          inputProps={{ 'aria-label': `Pilih ${row.label}` }}
        />
      ),
    },
  ];

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>Pilih role action</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <Typography color="text.secondary">Pilih satu atau lebih action yang akan diberikan ke role.</Typography>
          <SearchBox value={search} onChange={setSearch} placeholder="Cari kode atau nama action" />
          {error && <Alert severity="error">{error}</Alert>}
          {loading ? <Stack alignItems="center" sx={{ py: 4 }}><CircularProgress /></Stack> : (
            <ListComponent
              title="Role actions"
              columns={columns}
              rows={visibleActions}
              showSearch={false}
              showSortDirection={false}
              sortableFields={['sort_order', 'label']}
              initialSortField="sort_order"
              initialRowsPerPage={10}
              emptyMessage="Role action tidak ditemukan."
            />
          )}
          <FormControlLabel
            control={<Checkbox checked={visibleActions.length > 0 && visibleActions.every((item) => selected.includes(item.value))} indeterminate={visibleActions.some((item) => selected.includes(item.value)) && !visibleActions.every((item) => selected.includes(item.value))} onChange={(event) => setSelected((current) => event.target.checked ? [...new Set([...current, ...visibleActions.map((item) => item.value)])] : current.filter((value) => !visibleActions.some((item) => item.value === value)))} />}
            label={`${selected.length} action dipilih`}
          />
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Batal</Button>
        <Button variant="contained" onClick={() => onApply(actions.filter((action) => selected.includes(action.value)))}>Gunakan pilihan</Button>
      </DialogActions>
    </Dialog>
  );
}
