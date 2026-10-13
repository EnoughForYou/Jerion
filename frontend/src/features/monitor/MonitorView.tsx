import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const MonitorView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('monitor.title')}</CardTitle>
          <CardDescription>{t('monitor.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('monitor.noEvents')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
