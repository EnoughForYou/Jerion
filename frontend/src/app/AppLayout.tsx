import React, { useState, useEffect, useCallback } from 'react';
import { useLocation, useNavigate, Outlet } from 'react-router-dom';
import {
  Shield,
  Package,
  Search,
  Radar,
  ClipboardCheck,
  ShieldCheck,
  FileText,
  Settings,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  SidebarProvider,
  Sidebar,
  SidebarBody,
  SidebarLink,
  SidebarToggle,
} from '../components/aceternity/Sidebar';
import { DeviceHeader } from './DeviceHeader';
import { StatusBar } from './StatusBar';
import { getBridge } from '../lib/bridge';
import { useDeviceStore } from './deviceStore';

export const AppLayout: React.FC = () => {
  const { t } = useTranslation();
  const location = useLocation();
  const navigate = useNavigate();
  const { setActiveDevice } = useDeviceStore();
  const [isRefreshing, setIsRefreshing] = useState(false);

  const bridge = getBridge();

  const handleRefreshDevice = useCallback(async () => {
    try {
      setIsRefreshing(true);
      const dev = await bridge.device.getConnected();
      setActiveDevice(dev);
    } catch (err) {
      console.error('Failed to refresh device state:', err);
    } finally {
      setIsRefreshing(false);
    }
  }, [bridge, setActiveDevice]);

  useEffect(() => {
    let isMounted = true;

    async function initDevice() {
      try {
        const dev = await bridge.device.getConnected();
        if (isMounted) {
          setActiveDevice(dev);
        }
      } catch (err) {
        console.error('Initial device load failed:', err);
      }
    }

    initDevice();

    const unsub = bridge.on('device:changed', (device) => {
      if (isMounted) {
        setActiveDevice(device);
      }
    });

    return () => {
      isMounted = false;
      unsub();
    };
  }, [bridge, setActiveDevice]);

  const navItems = [
    { path: '/', label: t('nav.home'), icon: <Shield className="w-4 h-4" /> },
    { path: '/apps', label: t('nav.apps'), icon: <Package className="w-4 h-4" /> },
    { path: '/scan', label: t('nav.scan'), icon: <Search className="w-4 h-4" /> },
    { path: '/monitor', label: t('nav.monitor'), icon: <Radar className="w-4 h-4" /> },
    { path: '/audit', label: t('nav.audit'), icon: <ClipboardCheck className="w-4 h-4" /> },
    { path: '/remediate', label: t('nav.cleanup'), icon: <ShieldCheck className="w-4 h-4" /> },
    { path: '/reports', label: t('nav.reports'), icon: <FileText className="w-4 h-4" /> },
    { path: '/settings', label: t('nav.settings'), icon: <Settings className="w-4 h-4" /> },
  ];

  return (
    <SidebarProvider defaultOpen={true}>
      <div className="flex h-screen w-full min-w-0 min-h-0 overflow-hidden bg-base text-text-primary">
        {/* Left Sidebar Navigation */}
        <Sidebar>
          <SidebarBody>
            <div className="flex flex-col space-y-1">
              {/* App Brand Header */}
              <div className="flex items-center gap-2.5 px-2 py-3 mb-2 border-b border-separator/20">
                <div className="w-7 h-7 rounded-md bg-accent flex items-center justify-center text-white font-bold text-xs shadow-xs">
                  J
                </div>
                <div className="flex flex-col truncate">
                  <span className="font-semibold text-base leading-tight text-text-primary">
                    {t('app.title')}
                  </span>
                  <span className="text-[10px] text-text-secondary leading-none">
                    {t('app.tagline')}
                  </span>
                </div>
              </div>

              {/* Navigation Items */}
              {navItems.map((item) => (
                <SidebarLink
                  key={item.path}
                  icon={item.icon}
                  label={item.label}
                  active={location.pathname === item.path}
                  onClick={() => navigate(item.path)}
                />
              ))}
            </div>

            {/* Bottom Collapse Toggle */}
            <div className="pt-2 border-t border-separator/20">
              <SidebarToggle />
            </div>
          </SidebarBody>
        </Sidebar>

        {/* Main Content Area */}
        <div className="flex flex-col flex-1 h-full min-w-0 overflow-hidden">
          <DeviceHeader onRefresh={handleRefreshDevice} isRefreshing={isRefreshing} />

          <main className="flex-1 overflow-y-auto p-3 sm:p-5 md:p-6">
            <div className="max-w-[1200px] mx-auto w-full min-h-0 flex-1 flex flex-col">
              <Outlet />
            </div>
          </main>

          <StatusBar />
        </div>
      </div>
    </SidebarProvider>
  );
};
