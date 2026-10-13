import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const SettingsView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('settings.title')}</CardTitle>
          <CardDescription>{t('settings.general')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('settings.theme')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
