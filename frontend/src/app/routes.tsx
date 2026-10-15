import React from 'react';
import { createHashRouter, Navigate } from 'react-router-dom';
import { AppLayout } from './AppLayout';
import { HomeView } from '../features/home/HomeView';
import { AppsView } from '../features/apps/AppsView';
import { ScanView } from '../features/scan/ScanView';
import { MonitorView } from '../features/monitor/MonitorView';
import { AuditView } from '../features/audit/AuditView';
import { RemediateView } from '../features/remediate/RemediateView';
import { ReportsView } from '../features/reports/ReportsView';
import { SettingsView } from '../features/settings/SettingsView';
import { ErrorBoundary, RouteErrorBoundary } from '../components/ui/ErrorBoundary';

export const router = createHashRouter([
  {
    path: '/',
    element: (
      <ErrorBoundary>
        <AppLayout />
      </ErrorBoundary>
    ),
    errorElement: <RouteErrorBoundary />,
    children: [
      {
        index: true,
        element: (
          <ErrorBoundary>
            <HomeView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'apps',
        element: (
          <ErrorBoundary>
            <AppsView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'scan',
        element: (
          <ErrorBoundary>
            <ScanView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'monitor',
        element: (
          <ErrorBoundary>
            <MonitorView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'audit',
        element: (
          <ErrorBoundary>
            <AuditView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'remediate',
        element: (
          <ErrorBoundary>
            <RemediateView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'reports',
        element: (
          <ErrorBoundary>
            <ReportsView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      {
        path: 'settings',
        element: (
          <ErrorBoundary>
            <SettingsView />
          </ErrorBoundary>
        ),
        errorElement: <RouteErrorBoundary />,
      },
      { path: '*', element: <Navigate to="/" replace /> },
    ],
  },
]);
