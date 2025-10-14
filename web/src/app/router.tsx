import React from 'react';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import { ROUTES } from '../shared/constants/routes';
import DashboardPage from '../features/dashboard/DashboardPage';
import ReportsPage from '../features/reports/ReportsPage';
import ReportDetailPage from '../features/reports/ReportDetailPage';
import PoliciesPage from '../features/policies/PoliciesPage';

const router = createBrowserRouter([
  {
    path: ROUTES.DASHBOARD,
    element: <DashboardPage />,
  },
  {
    path: ROUTES.REPORTS,
    element: <ReportsPage />,
  },
  {
    path: ROUTES.REPORT_DETAIL,
    element: <ReportDetailPage />,
  },
  {
    path: ROUTES.POLICIES,
    element: <PoliciesPage />,
  },
  {
    path: '*',
    element: <DashboardPage />, // Redirect to dashboard for unknown routes
  },
]);

export const AppRouter: React.FC = () => {
  return <RouterProvider router={router} />;
};