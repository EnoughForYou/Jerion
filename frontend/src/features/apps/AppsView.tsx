import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const AppsView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('apps.title')}</CardTitle>
          <CardDescription>{t('apps.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('apps.searchPlaceholder')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
