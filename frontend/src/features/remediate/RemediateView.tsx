import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const RemediateView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('cleanup.title')}</CardTitle>
          <CardDescription>{t('cleanup.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('cleanup.noActions')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
