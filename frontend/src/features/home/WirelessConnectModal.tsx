import React, { useState } from 'react';
import {
  Wifi,
  KeyRound,
  Network,
  AlertCircle,
  CheckCircle2,
  Info,
  ArrowRight,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../components/ui/Dialog';
import { Button } from '../../components/ui/Button';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../../components/ui/Tabs';
import { getBridge } from '../../lib/bridge';

export interface WirelessConnectModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConnected?: (endpoint: string) => void;
}

export const WirelessConnectModal: React.FC<WirelessConnectModalProps> = ({
  open,
  onOpenChange,
  onConnected,
}) => {
  const { t } = useTranslation();
  const bridge = getBridge();

  const [activeTab, setActiveTab] = useState<'pair' | 'connect'>('pair');

  // Pair state
  const [pairHost, setPairHost] = useState('');
  const [pairingPort, setPairingPort] = useState('');
  const [pairingCode, setPairingCode] = useState('');
  const [isPairing, setIsPairing] = useState(false);
  const [pairError, setPairError] = useState<string | null>(null);
  const [pairSuccess, setPairSuccess] = useState<string | null>(null);

  // Connect state
  const [connectHost, setConnectHost] = useState('');
  const [connectPort, setConnectPort] = useState('5555');
  const [isConnecting, setIsConnecting] = useState(false);
  const [connectError, setConnectError] = useState<string | null>(null);
  const [connectSuccess, setConnectSuccess] = useState<string | null>(null);

  const resetForm = () => {
    setPairError(null);
    setPairSuccess(null);
    setConnectError(null);
    setConnectSuccess(null);
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      resetForm();
    }
    onOpenChange(nextOpen);
  };

  const handlePair = async (e: React.FormEvent) => {
    e.preventDefault();
    setPairError(null);
    setPairSuccess(null);

    const host = pairHost.trim();
    const port = parseInt(pairingPort.trim(), 10);
    const code = pairingCode.trim();

    if (!host) {
      setPairError('Please enter the device IP address');
      return;
    }
    if (isNaN(port) || port < 1 || port > 65535) {
      setPairError('Please enter a valid pairing port (1-65535)');
      return;
    }
    if (!/^\d{6}$/.test(code)) {
      setPairError('Pairing code must be exactly 6 digits');
      return;
    }

    try {
      setIsPairing(true);
      await bridge.device.wirelessPair(host, port, code);
      setPairSuccess(t('device.wireless.pairSuccess'));
      setConnectHost(host);
      // Auto-switch to connect tab after successful pairing
      setTimeout(() => {
        setActiveTab('connect');
      }, 1500);
    } catch (err: unknown) {
      setPairError(err instanceof Error ? err.message : 'Failed to pair device');
    } finally {
      setIsPairing(false);
    }
  };

  const handleConnect = async (e: React.FormEvent) => {
    e.preventDefault();
    setConnectError(null);
    setConnectSuccess(null);

    const host = (connectHost || pairHost).trim();
    const port = parseInt(connectPort.trim(), 10);

    if (!host) {
      setConnectError('Please enter the device IP address');
      return;
    }
    if (isNaN(port) || port < 1 || port > 65535) {
      setConnectError('Please enter a valid connection port (1-65535)');
      return;
    }

    try {
      setIsConnecting(true);
      const endpoint = await bridge.device.wirelessConnect(host, port);
      setConnectSuccess(t('device.wireless.connectSuccess', { endpoint }));
      if (onConnected) {
        onConnected(endpoint);
      }
      setTimeout(() => {
        handleOpenChange(false);
      }, 1200);
    } catch (err: unknown) {
      setConnectError(err instanceof Error ? err.message : 'Failed to connect wirelessly');
    } finally {
      setIsConnecting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-lg p-6">
        <DialogHeader className="pb-3 border-b border-separator-subtle">
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-md bg-accent/15 text-accent flex items-center justify-center shrink-0">
              <Wifi className="w-4 h-4" />
            </div>
            <div>
              <DialogTitle>{t('device.wireless.title')}</DialogTitle>
              <DialogDescription className="text-xs">
                {t('device.wireless.subtitle')}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="py-3 space-y-4">
          {/* Instructions Box */}
          <div className="p-3 rounded-md bg-raised/50 border border-separator-subtle space-y-1.5 text-xs">
            <div className="flex items-center gap-1.5 font-semibold text-text-primary">
              <Info className="w-3.5 h-3.5 text-accent shrink-0" />
              <span>{t('device.wireless.instructionsTitle')}</span>
            </div>
            <ol className="list-decimal list-inside space-y-0.5 text-text-secondary text-[11px] leading-relaxed pl-1">
              <li>{t('device.wireless.step1')}</li>
              <li>{t('device.wireless.step2')}</li>
              <li>{t('device.wireless.step3')}</li>
            </ol>
          </div>

          <Tabs
            value={activeTab}
            onValueChange={(val) => {
              setActiveTab(val as 'pair' | 'connect');
              resetForm();
            }}
          >
            <TabsList className="grid grid-cols-2 w-full">
              <TabsTrigger value="pair" className="flex items-center gap-1.5">
                <KeyRound className="w-3.5 h-3.5" />
                <span>{t('device.wireless.tabPair')}</span>
              </TabsTrigger>
              <TabsTrigger value="connect" className="flex items-center gap-1.5">
                <Network className="w-3.5 h-3.5" />
                <span>{t('device.wireless.tabConnect')}</span>
              </TabsTrigger>
            </TabsList>

            {/* TAB 1: PAIR */}
            <TabsContent value="pair" className="pt-2 space-y-3">
              {pairError && (
                <div className="p-2.5 rounded bg-sev-critical/10 border border-sev-critical/30 text-xs text-sev-critical flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{pairError}</span>
                </div>
              )}

              {pairSuccess && (
                <div className="p-2.5 rounded bg-ok/10 border border-ok/30 text-xs text-ok flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 shrink-0" />
                  <span>{pairSuccess}</span>
                </div>
              )}

              <form onSubmit={handlePair} className="space-y-3">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-text-primary block">
                    {t('device.wireless.hostLabel')}
                  </label>
                  <input
                    type="text"
                    value={pairHost}
                    onChange={(e) => setPairHost(e.target.value)}
                    placeholder={t('device.wireless.hostPlaceholder')}
                    className="w-full h-8 px-2.5 rounded-sm bg-base border border-separator-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:ring-1 focus:ring-accent font-mono"
                    required
                  />
                </div>

                <div className="grid grid-cols-2 gap-2">
                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-text-primary block">
                      {t('device.wireless.pairingPortLabel')}
                    </label>
                    <input
                      type="number"
                      value={pairingPort}
                      onChange={(e) => setPairingPort(e.target.value)}
                      placeholder={t('device.wireless.pairingPortPlaceholder')}
                      className="w-full h-8 px-2.5 rounded-sm bg-base border border-separator-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:ring-1 focus:ring-accent font-mono"
                      required
                    />
                    <span className="text-[10px] text-text-secondary block">
                      {t('device.wireless.pairingPortHint')}
                    </span>
                  </div>

                  <div className="space-y-1">
                    <label className="text-xs font-semibold text-text-primary block">
                      {t('device.wireless.pairingCodeLabel')}
                    </label>
                    <input
                      type="text"
                      maxLength={6}
                      value={pairingCode}
                      onChange={(e) => setPairingCode(e.target.value.replace(/\D/g, ''))}
                      placeholder={t('device.wireless.pairingCodePlaceholder')}
                      className="w-full h-8 px-2.5 rounded-sm bg-base border border-separator-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:ring-1 focus:ring-accent font-mono tracking-widest text-center"
                      required
                    />
                  </div>
                </div>

                <div className="pt-2 flex justify-end">
                  <Button
                    type="submit"
                    variant="primary"
                    size="sm"
                    disabled={isPairing}
                  >
                    <span>
                      {isPairing
                        ? t('device.wireless.pairing')
                        : t('device.wireless.pairAction')}
                    </span>
                    <ArrowRight className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </form>
            </TabsContent>

            {/* TAB 2: CONNECT */}
            <TabsContent value="connect" className="pt-2 space-y-3">
              {connectError && (
                <div className="p-2.5 rounded bg-sev-critical/10 border border-sev-critical/30 text-xs text-sev-critical flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 shrink-0" />
                  <span>{connectError}</span>
                </div>
              )}

              {connectSuccess && (
                <div className="p-2.5 rounded bg-ok/10 border border-ok/30 text-xs text-ok flex items-center gap-2">
                  <CheckCircle2 className="w-4 h-4 shrink-0" />
                  <span>{connectSuccess}</span>
                </div>
              )}

              <form onSubmit={handleConnect} className="space-y-3">
                <div className="space-y-1">
                  <label className="text-xs font-semibold text-text-primary block">
                    {t('device.wireless.hostLabel')}
                  </label>
                  <input
                    type="text"
                    value={connectHost || pairHost}
                    onChange={(e) => setConnectHost(e.target.value)}
                    placeholder={t('device.wireless.hostPlaceholder')}
                    className="w-full h-8 px-2.5 rounded-sm bg-base border border-separator-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:ring-1 focus:ring-accent font-mono"
                    required
                  />
                </div>

                <div className="space-y-1">
                  <label className="text-xs font-semibold text-text-primary block">
                    {t('device.wireless.connectionPortLabel')}
                  </label>
                  <input
                    type="number"
                    value={connectPort}
                    onChange={(e) => setConnectPort(e.target.value)}
                    placeholder={t('device.wireless.connectionPortPlaceholder')}
                    className="w-full h-8 px-2.5 rounded-sm bg-base border border-separator-subtle text-xs text-text-primary placeholder:text-text-secondary/50 focus:outline-none focus:ring-1 focus:ring-accent font-mono"
                    required
                  />
                  <span className="text-[10px] text-text-secondary block">
                    {t('device.wireless.connectionPortHint')}
                  </span>
                </div>

                <div className="pt-2 flex justify-end">
                  <Button
                    type="submit"
                    variant="primary"
                    size="sm"
                    disabled={isConnecting}
                  >
                    <span>
                      {isConnecting
                        ? t('device.wireless.connecting')
                        : t('device.wireless.connectAction')}
                    </span>
                    <Wifi className="w-3.5 h-3.5" />
                  </Button>
                </div>
              </form>
            </TabsContent>
          </Tabs>
        </div>

        <DialogFooter className="pt-2 border-t border-separator-subtle">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => handleOpenChange(false)}
          >
            {t('common.close')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
