import React from 'react';
import { useTranslation } from 'react-i18next';
import { Card, CardHeader, CardTitle, CardDescription, CardContent } from '../../components/ui/Card';

export const AuditView: React.FC = () => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle>{t('audit.title')}</CardTitle>
          <CardDescription>{t('audit.description')}</CardDescription>
        </CardHeader>
        <CardContent>
          <p className="text-text-secondary">{t('common.loading')}</p>
        </CardContent>
      </Card>
    </div>
  );
};
