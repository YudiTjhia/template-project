import { createTheme } from '@mui/material/styles';
import baseTheme from '@/app/themes/baseTheme';

const customerTheme = createTheme(baseTheme, {
  palette: {
    primary: { main: '#00897b' },
    secondary: { main: '#f9a825' },
  },
  shape: {
    borderRadius: 12,
  },
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 12,
        },
      },
    },
  },
});

export default customerTheme;
