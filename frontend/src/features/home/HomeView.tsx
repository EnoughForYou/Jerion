import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const HomeView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('home.title')}</CardTitle>
          <CardDescription>{t('home.deviceInfo')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('home.scanCtaSubtitle')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
