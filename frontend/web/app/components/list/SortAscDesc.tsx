'use client';

import ArrowDownwardIcon from '@mui/icons-material/ArrowDownward';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import { IconButton, Tooltip } from '@mui/material';

export type SortDirection = 'asc' | 'desc';
export type SortAscDescProps = {
  value: SortDirection;
  onChange: (value: SortDirection) => void;
  ascLabel?: string;
  descLabel?: string;
  disabled?: boolean;
};

export default function SortAscDesc({
  value,
  onChange,
  ascLabel = 'Urutkan naik',
  descLabel = 'Urutkan turun',
  disabled = false,
}: SortAscDescProps) {
  const nextValue = value === 'asc' ? 'desc' : 'asc';
  const label = value === 'asc' ? ascLabel : descLabel;

  return (
    <Tooltip title={`${label} (klik untuk mengubah)`}>
      <span>
        <IconButton
          aria-label={label}
          disabled={disabled}
          onClick={() => onChange(nextValue)}
          size="small"
        >
          {value === 'asc' ? <ArrowUpwardIcon fontSize="small" /> : <ArrowDownwardIcon fontSize="small" />}
        </IconButton>
      </span>
    </Tooltip>
  );
}
