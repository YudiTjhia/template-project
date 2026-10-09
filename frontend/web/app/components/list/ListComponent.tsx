'use client';

import { useMemo, useState } from 'react';
import AddIcon from '@mui/icons-material/Add';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import EditOutlinedIcon from '@mui/icons-material/EditOutlined';
import VisibilityOutlinedIcon from '@mui/icons-material/VisibilityOutlined';
import {
  Box,
  Button,
  IconButton,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TablePagination,
  TableRow,
} from '@mui/material';
import FilterField, { type FilterOption } from './FilterField';
import SearchBox from './SearchBox';
import SortAscDesc, { type SortDirection } from './SortAscDesc';
import SortingBy from './SortingBy';
import PageTitle from '../page/PageTitle';

export type ListRow = Record<string, unknown> & { id: string | number };

export type ListColumn<T extends ListRow> = {
  field: keyof T;
  label: string;
  searchable?: boolean;
  render?: (value: T[keyof T], row: T) => React.ReactNode;
};

export type ListFilter<T extends ListRow> = {
  field: keyof T;
  label: string;
  options: FilterOption[];
  allLabel?: string;
};

export type ListComponentProps<T extends ListRow> = {
  title: string;
  description?: string;
  columns: ListColumn<T>[];
  rows: T[];
  filters?: ListFilter<T>[];
  sortableFields?: (keyof T)[];
  showSearch?: boolean;
  showSortDirection?: boolean;
  addLabel?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  rowsPerPageOptions?: number[];
  initialRowsPerPage?: number;
  initialSortField?: keyof T;
  initialSortDirection?: SortDirection;
  getRowLabel?: (row: T) => string;
  onAdd?: () => void;
  onEdit?: (row: T) => void;
  onView?: (row: T) => void;
  onDelete?: (row: T) => void;
};

function compareValues(left: unknown, right: unknown) {
  if (left == null && right == null) return 0;
  if (left == null) return -1;
  if (right == null) return 1;
  if (typeof left === 'number' && typeof right === 'number') return left - right;

  const leftDate = typeof left === 'string' ? Date.parse(left) : NaN;
  const rightDate = typeof right === 'string' ? Date.parse(right) : NaN;
  if (!Number.isNaN(leftDate) && !Number.isNaN(rightDate)) return leftDate - rightDate;

  return String(left).localeCompare(String(right), 'id', { numeric: true, sensitivity: 'base' });
}

export default function ListComponent<T extends ListRow>({
  title,
  description,
  columns,
  rows,
  filters = [],
  sortableFields,
  showSearch = true,
  showSortDirection = true,
  addLabel = 'Tambah data',
  searchPlaceholder = 'Cari data',
  emptyMessage = 'Data tidak ditemukan.',
  rowsPerPageOptions = [5, 10, 25],
  initialRowsPerPage = 10,
  initialSortField,
  initialSortDirection = 'asc',
  getRowLabel = (row) => String(row.id),
  onAdd,
  onEdit,
  onView,
  onDelete,
}: ListComponentProps<T>) {
  const [query, setQuery] = useState('');
  const [filterValues, setFilterValues] = useState<Record<string, string>>({});
  const [sortField, setSortField] = useState<string>(String(initialSortField ?? sortableFields?.[0] ?? ''));
  const [sortDirection, setSortDirection] = useState<SortDirection>(initialSortDirection);
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(initialRowsPerPage);
  const searchableColumns = columns.filter((column) => column.searchable !== false);
  const activeSortableFields = sortableFields ?? columns.map((column) => column.field);
  const sortOptions = activeSortableFields.map((field) => ({
    value: String(field),
    label: columns.find((column) => column.field === field)?.label ?? String(field),
  }));

  const filteredRows = useMemo(() => {
    const normalizedQuery = query.trim().toLocaleLowerCase('id');
    const searched = rows.filter((row) => {
      const matchesQuery = !normalizedQuery || searchableColumns.some((column) =>
        String(row[column.field] ?? '').toLocaleLowerCase('id').includes(normalizedQuery),
      );
      const matchesFilters = filters.every((filter) => {
        const selectedValue = filterValues[String(filter.field)];
        return !selectedValue || String(row[filter.field] ?? '') === selectedValue;
      });
      return matchesQuery && matchesFilters;
    });

    if (!sortField) return searched;
    return [...searched].sort((left, right) => {
      const result = compareValues(left[sortField], right[sortField]);
      return sortDirection === 'asc' ? result : -result;
    });
  }, [rows, query, filterValues, filters, searchableColumns, sortField, sortDirection]);

  const visibleRows = filteredRows.slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage);
  const hasActions = Boolean(onView || onEdit || onDelete);
  const columnCount = columns.length + Number(hasActions);

  return (
    <Box sx={{ width: '100%', maxWidth: 'none', mx: 0, p: { xs: 2, md: 4 } }}>
      <PageTitle
        title={title}
        subtitle={description}
        actions={onAdd ? (
          <Button onClick={onAdd} variant="contained" startIcon={<AddIcon />}>
            {addLabel}
          </Button>
        ) : undefined}
      />

      <Paper variant="outlined" sx={{ overflow: 'hidden', borderRadius: 2 }}>
        <Stack direction={{ xs: 'column', md: 'row' }} spacing={1.5} useFlexGap flexWrap="wrap" sx={{ p: 2 }}>
          {showSearch && <SearchBox value={query} onChange={(value) => { setQuery(value); setPage(0); }} placeholder={searchPlaceholder} />}
          {filters.map((filter) => (
            <FilterField
              key={String(filter.field)}
              label={filter.label}
              options={filter.options}
              allLabel={filter.allLabel}
              value={filterValues[String(filter.field)] ?? ''}
              onChange={(value) => {
                setFilterValues((current) => ({ ...current, [String(filter.field)]: value }));
                setPage(0);
              }}
            />
          ))}
          {sortOptions.length > 0 && (
            <Stack direction="row" alignItems="center" spacing={0.5}>
              <SortingBy
                value={sortField}
                options={sortOptions}
                onChange={(value) => { setSortField(value); setPage(0); }}
              />
              {showSortDirection && <SortAscDesc value={sortDirection} onChange={setSortDirection} />}
            </Stack>
          )}
        </Stack>

        <TableContainer>
          <Table aria-label={title}>
            <TableHead>
              <TableRow>
                {columns.map((column) => <TableCell key={String(column.field)}>{column.label}</TableCell>)}
                {hasActions && <TableCell align="right">Aksi</TableCell>}
              </TableRow>
            </TableHead>
            <TableBody>
              {visibleRows.map((row) => {
                const rowLabel = getRowLabel(row);
                return (
                  <TableRow hover key={row.id}>
                    {columns.map((column) => {
                      const value = row[column.field] as T[keyof T];
                      return <TableCell key={String(column.field)}>{column.render ? column.render(value, row) : String(value ?? '')}</TableCell>;
                    })}
                    {hasActions && (
                      <TableCell align="right" sx={{ whiteSpace: 'nowrap' }}>
                        {onEdit && <IconButton aria-label={`Edit ${rowLabel}`} onClick={() => onEdit(row)} size="small"><EditOutlinedIcon fontSize="small" /></IconButton>}
                        {onView && <IconButton aria-label={`Lihat ${rowLabel}`} onClick={() => onView(row)} size="small"><VisibilityOutlinedIcon fontSize="small" /></IconButton>}
                        {onDelete && <IconButton aria-label={`Hapus ${rowLabel}`} onClick={() => onDelete(row)} size="small" color="error"><DeleteOutlineIcon fontSize="small" /></IconButton>}
                      </TableCell>
                    )}
                  </TableRow>
                );
              })}
              {visibleRows.length === 0 && (
                <TableRow><TableCell colSpan={columnCount} align="center" sx={{ py: 6, color: 'text.secondary' }}>{emptyMessage}</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          component="div"
          count={filteredRows.length}
          page={page}
          onPageChange={(_, nextPage) => setPage(nextPage)}
          rowsPerPage={rowsPerPage}
          onRowsPerPageChange={(event) => { setRowsPerPage(Number(event.target.value)); setPage(0); }}
          rowsPerPageOptions={rowsPerPageOptions}
          labelRowsPerPage="Baris per halaman:"
          labelDisplayedRows={({ from, to, count }) => `${from}–${to} dari ${count}`}
        />
      </Paper>
    </Box>
  );
}
