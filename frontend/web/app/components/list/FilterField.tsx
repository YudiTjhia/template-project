'use client';

import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';

export type FilterOption = { label: string; value: string };
export type FilterFieldProps = {
  label: string;
  value: string;
  options: FilterOption[];
  onChange: (value: string) => void;
  allLabel?: string;
  allValue?: string;
  width?: number | string;
};

export default function FilterField({
  label,
  value,
  options,
  onChange,
  allLabel = 'Semua',
  allValue = '',
  width = 200,
}: FilterFieldProps) {
  const labelId = `${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}-filter-label`;

  return (
    <FormControl size="small" sx={{ minWidth: width }}>
      <InputLabel id={labelId}>{label}</InputLabel>
      <Select labelId={labelId} value={value} label={label} onChange={(event) => onChange(event.target.value)}>
        <MenuItem value={allValue}>{allLabel}</MenuItem>
        {options.map((option) => (
          <MenuItem key={option.value} value={option.value}>{option.label}</MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}
