'use client';

import SearchIcon from '@mui/icons-material/Search';
import { InputAdornment, TextField } from '@mui/material';

export type SearchBoxProps = {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  label?: string;
  fullWidth?: boolean;
};

export default function SearchBox({
  value,
  onChange,
  placeholder = 'Cari data',
  label,
  fullWidth = false,
}: SearchBoxProps) {
  return (
    <TextField
      value={value}
      onChange={(event) => onChange(event.target.value)}
      placeholder={placeholder}
      aria-label={label ?? placeholder}
      size="small"
      fullWidth={fullWidth}
      sx={fullWidth ? undefined : { width: { xs: '100%', sm: 360 } }}
      InputProps={{
        startAdornment: <InputAdornment position="start"><SearchIcon fontSize="small" /></InputAdornment>,
      }}
    />
  );
}
