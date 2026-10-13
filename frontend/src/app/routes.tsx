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

export const router = createHashRouter([
  {
    path: '/',
    element: <AppLayout />,
    children: [
      { index: true, element: <HomeView /> },
      { path: 'apps', element: <AppsView /> },
      { path: 'scan', element: <ScanView /> },
      { path: 'monitor', element: <MonitorView /> },
      { path: 'audit', element: <AuditView /> },
      { path: 'remediate', element: <RemediateView /> },
      { path: 'reports', element: <ReportsView /> },
      { path: 'settings', element: <SettingsView /> },
      { path: '*', element: <Navigate to="/" replace /> },
    ],
  },
]);
