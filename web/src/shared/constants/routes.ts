export const ROUTES = {
  DASHBOARD: '/',
  REPORTS: '/reports',
  REPORT_DETAIL: '/reports/:id',
  POLICIES: '/policies',
  POLICY_DETAIL: '/policies/:id',
} as const;

export const NAV_ITEMS = [
  {
    path: ROUTES.DASHBOARD,
    label: 'Dashboard',
    icon: 'BarChart3',
  },
  {
    path: ROUTES.REPORTS,
    label: 'Reports',
    icon: 'FileText',
  },
  {
    path: ROUTES.POLICIES,
    label: 'Policies',
    icon: 'Settings',
  },
] as const;