import React from 'react';
import { Shield, AlertCircle, FileText, Terminal, ArrowRight } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../../components/ui/Dialog';
import { Badge } from '../../components/ui/Badge';
import { Button } from '../../components/ui/Button';
import type { Finding } from '../../lib/bridge/schema';

export interface EvidenceDrawerProps {
  finding: Finding | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export const EvidenceDrawer: React.FC<EvidenceDrawerProps> = ({
  finding,
  open,
  onOpenChange,
}) => {
  const { t } = useTranslation();
  const navigate = useNavigate();

  if (!finding) return null;

  const handleRemediate = () => {
    onOpenChange(false);
    navigate('/remediate');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-center gap-2 mb-1">
            <Badge severity={finding.severity} />
            <span className="text-xs font-mono text-text-secondary bg-raised px-2 py-0.5 rounded border border-separator/30">
              {finding.ruleId}
            </span>
            <span className="text-xs text-text-secondary">
              {t('scan.confidenceLabel', { confidence: finding.confidence.toUpperCase() })}
            </span>
          </div>
          <DialogTitle className="text-base font-semibold text-text-primary">
            {finding.title}
          </DialogTitle>
          <DialogDescription className="text-xs text-text-secondary font-mono">
            {finding.package
              ? t('scan.packageLabel', { package: finding.package })
              : 'System-level target'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2 text-xs">
          {/* Why Flagged */}
          <div className="p-3 rounded-md bg-raised/40 border border-separator/25 space-y-1.5">
            <div className="flex items-center gap-1.5 font-semibold text-text-primary">
              <Shield className="w-3.5 h-3.5 text-accent" />
              <span>{t('scan.evidenceDrawer.whyFlagged')}</span>
            </div>
            <p className="text-text-secondary text-[11px] leading-relaxed">
              {finding.whyFlagged}
            </p>
          </div>

          {/* False Positive Guidance */}
          <div className="p-3 rounded-md bg-surface border border-separator/25 space-y-1.5">
            <div className="flex items-center gap-1.5 font-semibold text-text-primary">
              <AlertCircle className="w-3.5 h-3.5 text-sev-medium" />
              <span>{t('scan.evidenceDrawer.falsePositiveContext')}</span>
            </div>
            <p className="text-text-secondary text-[11px] leading-relaxed">
              {finding.falsePositiveHint}
            </p>
          </div>

          {/* Evidence List */}
          <div className="space-y-2.5">
            <h4 className="text-xs font-semibold text-text-primary flex items-center gap-1.5">
              <FileText className="w-3.5 h-3.5 text-accent" />
              <span>
                {t('scan.evidenceDrawer.evidenceList', { count: finding.evidence?.length ?? 0 })}
              </span>
            </h4>

            {(finding.evidence || []).map((item, idx) => (
              <div
                key={idx}
                className="p-3 rounded-md bg-raised/20 border border-separator/20 space-y-2"
              >
                <div className="flex items-center justify-between">
                  <span className="font-semibold text-text-primary text-[11px]">
                    {item.title}
                  </span>
                  <span className="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-raised text-text-secondary border border-separator/20">
                    {item.type}
                  </span>
                </div>
                <p className="text-text-secondary text-[11px] leading-relaxed">
                  {item.details}
                </p>

                {item.rawOutput && (
                  <div className="space-y-1 pt-1">
                    <div className="flex items-center gap-1 text-[10px] text-text-secondary font-mono">
                      <Terminal className="w-3 h-3" />
                      <span>{t('scan.evidenceDrawer.rawOutput')}</span>
                    </div>
                    <pre className="p-2 rounded bg-base/80 border border-separator/20 text-[11px] font-mono text-text-primary overflow-x-auto whitespace-pre-wrap select-all">
                      {item.rawOutput}
                    </pre>
                  </div>
                )}
              </div>
            ))}
          </div>

          {/* Recommended Action */}
          {finding.recommendedAction && (
            <div className="p-3 rounded-md bg-accent/5 border border-accent/20 space-y-2">
              <div className="font-semibold text-accent text-xs">
                {t('scan.evidenceDrawer.recommendedAction')}: {finding.recommendedAction.title}
              </div>
              <p className="text-text-secondary text-[11px]">
                {t('scan.evidenceDrawer.impact', { impact: finding.recommendedAction.impact })}
              </p>
            </div>
          )}
        </div>

        <DialogFooter className="flex items-center justify-between sm:justify-between w-full">
          <Button
            variant="ghost"
            size="sm"
            onClick={() => onOpenChange(false)}
          >
            {t('scan.evidenceDrawer.close')}
          </Button>

          {finding.recommendedAction && (
            <Button
              variant="primary"
              size="sm"
              onClick={handleRemediate}
            >
              <span>{t('scan.evidenceDrawer.stageForCleanup')}</span>
              <ArrowRight className="w-3.5 h-3.5" />
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
