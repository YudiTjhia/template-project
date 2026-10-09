'use client';

import { FormControl, InputLabel, MenuItem, Select } from '@mui/material';

export type SortingOption = { label: string; value: string };
export type SortingByProps = {
  value: string;
  options: SortingOption[];
  onChange: (value: string) => void;
  label?: string;
  width?: number | string;
};

export default function SortingBy({
  value,
  options,
  onChange,
  label = 'Urutkan berdasarkan',
  width = 220,
}: SortingByProps) {
  const labelId = 'sorting-by-label';

  return (
    <FormControl size="small" sx={{ minWidth: width }}>
      <InputLabel id={labelId}>{label}</InputLabel>
      <Select labelId={labelId} value={value} label={label} onChange={(event) => onChange(event.target.value)}>
        {options.map((option) => (
          <MenuItem key={option.value} value={option.value}>{option.label}</MenuItem>
        ))}
      </Select>
    </FormControl>
  );
}
