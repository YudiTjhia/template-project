'use client';

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
} from '@mui/material';
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react';

export type ConfirmBoxOptions = {
  title?: string;
  message: ReactNode;
  confirmLabel?: string;
  cancelLabel?: string;
  confirmColor?: 'primary' | 'error' | 'warning' | 'success';
};

type PendingConfirmation = ConfirmBoxOptions & {
  resolve: (confirmed: boolean) => void;
};

type ConfirmBoxContextValue = {
  confirm: (options: ConfirmBoxOptions) => Promise<boolean>;
};

const ConfirmBoxContext = createContext<ConfirmBoxContextValue | null>(null);

export function ConfirmBoxProvider({ children }: { children: ReactNode }) {
  const [pending, setPending] = useState<PendingConfirmation | null>(null);
  const pendingRef = useRef<PendingConfirmation | null>(null);

  const finish = useCallback((confirmed: boolean) => {
    pendingRef.current?.resolve(confirmed);
    pendingRef.current = null;
    setPending(null);
  }, []);

  const confirm = useCallback((options: ConfirmBoxOptions) => {
    return new Promise<boolean>((resolve) => {
      // Resolve a previous prompt as cancelled if another one is opened.
      pendingRef.current?.resolve(false);
      const nextPending = { ...options, resolve };
      pendingRef.current = nextPending;
      setPending(nextPending);
    });
  }, []);

  const value = useMemo(() => ({ confirm }), [confirm]);

  return (
    <ConfirmBoxContext.Provider value={value}>
      {children}
      <Dialog open={Boolean(pending)} onClose={() => finish(false)} fullWidth maxWidth="xs">
        {pending && (
          <>
            <DialogTitle>{pending.title ?? 'Konfirmasi'}</DialogTitle>
            <DialogContent>
              <DialogContentText component="div">{pending.message}</DialogContentText>
            </DialogContent>
            <DialogActions>
              <Button onClick={() => finish(false)}>{pending.cancelLabel ?? 'Batal'}</Button>
              <Button
                variant="contained"
                color={pending.confirmColor ?? 'primary'}
                onClick={() => finish(true)}
                autoFocus
              >
                {pending.confirmLabel ?? 'Ya, lanjutkan'}
              </Button>
            </DialogActions>
          </>
        )}
      </Dialog>
    </ConfirmBoxContext.Provider>
  );
}

export function useConfirmBox() {
  const context = useContext(ConfirmBoxContext);
  if (!context) throw new Error('useConfirmBox must be used within ConfirmBoxProvider');
  return context;
}
