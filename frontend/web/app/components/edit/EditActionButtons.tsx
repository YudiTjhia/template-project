'use client';

import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import CancelOutlinedIcon from '@mui/icons-material/CancelOutlined';
import DeleteOutlineIcon from '@mui/icons-material/DeleteOutline';
import SaveOutlinedIcon from '@mui/icons-material/SaveOutlined';
import UpdateOutlinedIcon from '@mui/icons-material/UpdateOutlined';
import { Button, Stack } from '@mui/material';

export type EditActionButtonsProps = {
  showBack?: boolean;
  showCancel?: boolean;
  showSaveNew?: boolean;
  showUpdate?: boolean;
  showDelete?: boolean;
  onBack?: () => void;
  onCancel?: () => void;
  onSaveNew?: () => void;
  onUpdate?: () => void;
  onDelete?: () => void;
  backLabel?: string;
  cancelLabel?: string;
  saveNewLabel?: string;
  updateLabel?: string;
  deleteLabel?: string;
  disabled?: boolean;
  loading?: boolean;
  direction?: 'row' | 'column';
  align?: 'left' | 'center' | 'right';
};

export default function EditActionButtons({
  showBack = true,
  showCancel = true,
  showSaveNew = true,
  showUpdate = true,
  showDelete = true,
  onBack,
  onCancel,
  onSaveNew,
  onUpdate,
  onDelete,
  backLabel = 'Kembali',
  cancelLabel = 'Batal',
  saveNewLabel = 'Simpan baru',
  updateLabel = 'Perbarui',
  deleteLabel = 'Hapus',
  disabled = false,
  loading = false,
  direction = 'row',
  align = 'right',
}: EditActionButtonsProps) {
  const justifyContent = align === 'left' ? 'flex-start' : align === 'center' ? 'center' : 'flex-end';
  const buttonsDisabled = disabled || loading;

  return (
    <Stack
      direction={direction}
      spacing={1}
      useFlexGap
      flexWrap="wrap"
      justifyContent={justifyContent}
      alignItems={direction === 'row' ? 'center' : 'stretch'}
    >
      {showBack && (
        <Button variant="text" startIcon={<ArrowBackIcon />} onClick={onBack} disabled={buttonsDisabled}>
          {backLabel}
        </Button>
      )}
      {showCancel && (
        <Button variant="outlined" startIcon={<CancelOutlinedIcon />} onClick={onCancel} disabled={buttonsDisabled}>
          {cancelLabel}
        </Button>
      )}
      {showDelete && (
        <Button color="error" variant="outlined" startIcon={<DeleteOutlineIcon />} onClick={onDelete} disabled={buttonsDisabled}>
          {deleteLabel}
        </Button>
      )}
      {showSaveNew && (
        <Button variant="contained" startIcon={<SaveOutlinedIcon />} onClick={onSaveNew} disabled={buttonsDisabled}>
          {loading ? 'Menyimpan…' : saveNewLabel}
        </Button>
      )}
      {showUpdate && (
        <Button variant="contained" startIcon={<UpdateOutlinedIcon />} onClick={onUpdate} disabled={buttonsDisabled}>
          {loading ? 'Memperbarui…' : updateLabel}
        </Button>
      )}
    </Stack>
  );
}
