'use client';

import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import type { DialogProps } from '@mui/material/Dialog';
import ListComponent, { type ListComponentProps, type ListRow } from './ListComponent';

export type ListDialogProps<T extends ListRow> = ListComponentProps<T> & {
  open: boolean;
  onClose: () => void;
  maxWidth?: DialogProps['maxWidth'];
  fullWidth?: boolean;
  fullScreen?: boolean;
  closeOnBackdropClick?: boolean;
  closeOnEscape?: boolean;
};

export default function ListDialog<T extends ListRow>({
  open,
  onClose,
  maxWidth = 'lg',
  fullWidth = true,
  fullScreen = false,
  closeOnBackdropClick = true,
  closeOnEscape = true,
  ...listProps
}: ListDialogProps<T>) {
  const handleClose: NonNullable<DialogProps['onClose']> = (_event, reason) => {
    if (reason === 'backdropClick' && !closeOnBackdropClick) return;
    if (reason === 'escapeKeyDown' && !closeOnEscape) return;
    onClose();
  };

  return (
    <Dialog
      open={open}
      onClose={handleClose}
      maxWidth={maxWidth}
      fullWidth={fullWidth}
      fullScreen={fullScreen}
      aria-labelledby="list-dialog-title"
    >
      <DialogContent sx={{ p: 0 }}>
        <ListComponent {...listProps} />
      </DialogContent>
    </Dialog>
  );
}
