import { createTheme } from '@mui/material/styles';
import baseTheme from '@/app/themes/baseTheme';

const adminTheme = createTheme(baseTheme, {
  palette: {
    primary: { main: '#1976d2' },
    secondary: { main: '#455a64' },
  },
});

export default adminTheme;
