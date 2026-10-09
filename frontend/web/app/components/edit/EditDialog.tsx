'use client';

import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import type { DialogProps } from '@mui/material/Dialog';
import EditComponent, { type EditComponentProps } from './EditComponent';

export type EditDialogProps = Omit<EditComponentProps, 'onBack'> & {
  open: boolean;
  onClose: () => void;
  onBack?: EditComponentProps['onBack'];
  maxWidth?: DialogProps['maxWidth'];
  fullWidth?: boolean;
  fullScreen?: boolean;
  /** Close the dialog when the user clicks outside it. Defaults to true. */
  closeOnBackdropClick?: boolean;
  /** Close the dialog when the user presses Escape. Defaults to true. */
  closeOnEscape?: boolean;
};

export default function EditDialog({
  open,
  onClose,
  maxWidth = 'md',
  fullWidth = true,
  fullScreen = false,
  closeOnBackdropClick = true,
  closeOnEscape = true,
  onBack,
  onCancel,
  showBackButton,
  ...editProps
}: EditDialogProps) {
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
      aria-labelledby="edit-dialog-title"
    >
      <DialogContent sx={{ p: 0 }}>
        <EditComponent
          {...editProps}
          showBackButton={showBackButton}
          onBack={onBack ?? onClose}
          onCancel={onCancel ?? onClose}
        />
      </DialogContent>
    </Dialog>
  );
}
