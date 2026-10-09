'use client';

import { useState, type ReactNode } from 'react';
import {
  Box,
  Collapse,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Tooltip,
  useMediaQuery,
} from '@mui/material';
import { useTheme } from '@mui/material/styles';
import DashboardOutlinedIcon from '@mui/icons-material/DashboardOutlined';
import NotificationsActiveOutlinedIcon from '@mui/icons-material/NotificationsActiveOutlined';
import PeopleOutlineIcon from '@mui/icons-material/PeopleOutline';
import SecurityOutlinedIcon from '@mui/icons-material/SecurityOutlined';
import AccountCircleOutlinedIcon from '@mui/icons-material/AccountCircleOutlined';
import LockOutlinedIcon from '@mui/icons-material/LockOutlined';
import ManageAccountsOutlinedIcon from '@mui/icons-material/ManageAccountsOutlined';
import PersonOutlineIcon from '@mui/icons-material/PersonOutline';
import LogoutOutlinedIcon from '@mui/icons-material/LogoutOutlined';
import MenuOutlinedIcon from '@mui/icons-material/MenuOutlined';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import { usePathname, useRouter } from 'next/navigation';
import { getSubAppUrl, resolveSubAppType, type SubAppType } from '@/app/helpers/subApp';
import CurrentUser from './CurrentUser';

export type NavigationMenuItem = {
  label: string;
  href?: string;
  icon?: ReactNode;
  children?: NavigationMenuItem[];
  action?: 'logout';
};

export type NavigationMenuProps = {
  side?: 'left' | 'right';
  items?: NavigationMenuItem[];
  width?: number;
  onNavigate?: (item: NavigationMenuItem) => void;
  onLogout?: () => void;
  subAppType?: SubAppType;
  showCurrentUser?: boolean;
};

const defaultItems: NavigationMenuItem[] = [
  { label: 'Dashboard', href: '/', icon: <DashboardOutlinedIcon /> },
  { label: 'Notifications', href: '/notifications', icon: <NotificationsActiveOutlinedIcon /> },
  {
    label: 'User Management',
    icon: <PeopleOutlineIcon />,
    children: [
      { label: 'Users', href: '/users', icon: <PeopleOutlineIcon /> },
      { label: 'Roles', href: '/roles', icon: <SecurityOutlinedIcon /> },
    ],
  },
  {
    label: 'Profile',
    icon: <AccountCircleOutlinedIcon />,
    children: [
      { label: 'Change Password', href: '/profile/change-password', icon: <LockOutlinedIcon /> },
      { label: 'Account Profile', href: '/profile/account', icon: <ManageAccountsOutlinedIcon /> },
      { label: 'User Profile', href: '/profile/user', icon: <PersonOutlineIcon /> },
    ],
  },
  { label: 'Logout', href: '/logout', icon: <LogoutOutlinedIcon />, action: 'logout' },
];

export default function NavigationMenu({
  side = 'left',
  items = defaultItems,
  width = 264,
  onNavigate,
  onLogout,
  subAppType = resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE),
  showCurrentUser = true,
}: NavigationMenuProps) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const pathname = usePathname();
  const router = useRouter();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({});

  const navigate = (item: NavigationMenuItem) => {
    onNavigate?.(item);
    if (item.action === 'logout') {
      if (onLogout) onLogout();
      else router.push(getSubAppUrl(subAppType, item.href ?? '/logout'));
      setMobileOpen(false);
      return;
    }
    if (item.href) router.push(getSubAppUrl(subAppType, item.href));
    setMobileOpen(false);
  };

  const isSelected = (item: NavigationMenuItem) =>
    Boolean(item.href && (pathname === getSubAppUrl(subAppType, item.href) || (item.href !== '/' && pathname.startsWith(`${getSubAppUrl(subAppType, item.href)}/`))));

  const hasSelectedChild = (item: NavigationMenuItem) =>
    item.children?.some((child) => isSelected(child)) ?? false;

  const renderItems = (menuItems: NavigationMenuItem[], nested = false) => menuItems.map((item) => {
    const hasChildren = Boolean(item.children?.length);
    const open = expanded[item.label] ?? hasSelectedChild(item);
    const selected = isSelected(item) || hasSelectedChild(item);

    return (
      <Box key={item.label}>
        <ListItemButton
          selected={selected}
          onClick={() => {
            if (hasChildren) setExpanded((current) => ({ ...current, [item.label]: !open }));
            else navigate(item);
          }}
          sx={{ pl: nested ? 4 : 2, pr: 1.5, minHeight: 44, borderRadius: 1, mx: 1, my: 0.25 }}
        >
          {item.icon && <ListItemIcon sx={{ minWidth: 38, color: 'inherit' }}>{item.icon}</ListItemIcon>}
          <ListItemText primary={item.label} primaryTypographyProps={{ variant: 'body2', fontWeight: selected ? 600 : 400 }} />
          {hasChildren && (open ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />)}
        </ListItemButton>
        {hasChildren && (
          <Collapse in={open} timeout="auto" unmountOnExit>
            <List disablePadding>{renderItems(item.children!, true)}</List>
          </Collapse>
        )}
      </Box>
    );
  });

  const drawerContent = (
    <Box sx={{ width, height: '100%', display: 'flex', flexDirection: 'column' }}>
      <Toolbar>
        <Box component="span" sx={{ fontWeight: 700, color: 'primary.main' }}>Navigation</Box>
      </Toolbar>
      {showCurrentUser && <CurrentUser />}
      <Divider />
      <List sx={{ py: 1, flex: 1 }}>{renderItems(items)}</List>
    </Box>
  );

  if (isMobile) {
    return (
      <>
        <Tooltip title="Buka menu">
          <IconButton
            aria-label="Buka menu navigasi"
            onClick={() => setMobileOpen(true)}
            sx={{ position: 'fixed', zIndex: theme.zIndex.drawer + 1, top: 8, [side]: 8 }}
          >
            <MenuOutlinedIcon />
          </IconButton>
        </Tooltip>
        <Drawer
          anchor={side}
          variant="temporary"
          open={mobileOpen}
          onClose={() => setMobileOpen(false)}
          ModalProps={{ keepMounted: true }}
          sx={{ '& .MuiDrawer-paper': { width, boxSizing: 'border-box' } }}
        >
          {drawerContent}
        </Drawer>
      </>
    );
  }

  return (
    <Drawer
      anchor={side}
      variant="permanent"
      sx={{
        width,
        flexShrink: 0,
        '& .MuiDrawer-paper': { width, boxSizing: 'border-box', position: 'relative', height: '100vh' },
      }}
    >
      {drawerContent}
    </Drawer>
  );
}
