'use client';

import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
} from '@mui/material';
import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  useState,
  type ReactNode,
} from 'react';

export type MsgBoxOptions = {
  title?: string;
  message: ReactNode;
  severity?: 'success' | 'info' | 'warning' | 'error';
  closeLabel?: string;
};

type MsgBoxContextValue = {
  showMessage: (options: MsgBoxOptions) => void;
  closeMessage: () => void;
};

const MsgBoxContext = createContext<MsgBoxContextValue | null>(null);

export function MsgBoxProvider({ children }: { children: ReactNode }) {
  const [options, setOptions] = useState<MsgBoxOptions | null>(null);
  const closeMessage = useCallback(() => setOptions(null), []);
  const showMessage = useCallback((nextOptions: MsgBoxOptions) => setOptions(nextOptions), []);

  const value = useMemo(() => ({ showMessage, closeMessage }), [showMessage, closeMessage]);

  return (
    <MsgBoxContext.Provider value={value}>
      {children}
      <Dialog open={Boolean(options)} onClose={closeMessage} fullWidth maxWidth="xs">
        {options && (
          <>
            {options.title && <DialogTitle>{options.title}</DialogTitle>}
            <DialogContent>
              <Alert severity={options.severity ?? 'info'} variant="outlined">
                {options.message}
              </Alert>
            </DialogContent>
            <DialogActions>
              <Button onClick={closeMessage} autoFocus>
                {options.closeLabel ?? 'OK'}
              </Button>
            </DialogActions>
          </>
        )}
      </Dialog>
    </MsgBoxContext.Provider>
  );
}

export function useMsgBox() {
  const context = useContext(MsgBoxContext);
  if (!context) throw new Error('useMsgBox must be used within MsgBoxProvider');
  return context;
}
