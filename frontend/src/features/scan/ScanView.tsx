import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const ScanView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('scan.title')}</CardTitle>
          <CardDescription>{t('scan.quickScanDesc')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('scan.noFindings')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
