import React from 'react';
import { CheckCircle2, Loader2, Circle } from 'lucide-react';
import type { ScanProgress } from '../../lib/bridge/schema';

export interface ScanTimelineProps {
  progress: ScanProgress | null;
  isScanning: boolean;
}

interface StepItem {
  id: string;
  label: string;
  threshold: number;
}

const SCAN_STEPS: StepItem[] = [
  { id: 'manifests', label: 'Inspecting package manifests & signatures', threshold: 25 },
  { id: 'overlays', label: 'Analyzing overlay & background receivers', threshold: 60 },
  { id: 'system', label: 'Checking system-level configurations & ports', threshold: 90 },
  { id: 'summary', label: 'Evaluating threat rules & security summary', threshold: 100 },
];

export const ScanTimeline: React.FC<ScanTimelineProps> = ({
  progress,
  isScanning,
}) => {
  const currentPercentage = progress?.progress ?? 0;

  return (
    <div className="space-y-3 p-4 rounded-lg bg-surface border border-separator/35 shadow-sm text-xs">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {isScanning && <Loader2 className="w-4 h-4 text-accent animate-spin" />}
          <span className="font-semibold text-text-primary text-xs">
            {progress?.step || (isScanning ? 'Scan in progress...' : 'Scan Complete')}
          </span>
        </div>
        <span className="font-mono text-xs text-text-secondary font-medium">
          {Math.round(currentPercentage)}%
        </span>
      </div>

      {/* Progress bar */}
      <div className="h-1.5 w-full bg-raised rounded-full overflow-hidden">
        <div
          className="h-full bg-accent transition-all duration-300 ease-out"
          style={{ width: `${currentPercentage}%` }}
        />
      </div>

      {progress?.currentPackage && isScanning && (
        <div className="text-[11px] font-mono text-text-secondary truncate bg-raised/50 px-2 py-1 rounded border border-separator/20">
          Target: <span className="text-text-primary">{progress.currentPackage}</span>
        </div>
      )}

      {/* Step nodes */}
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2 pt-2 border-t border-separator/20">
        {SCAN_STEPS.map((step) => {
          const isDone = currentPercentage >= step.threshold || (!isScanning && currentPercentage > 0);
          const isCurrent = isScanning && currentPercentage < step.threshold && (
            SCAN_STEPS.find((s) => s.threshold > currentPercentage)?.id === step.id
          );

          return (
            <div
              key={step.id}
              className={`flex items-start gap-2 p-2 rounded border transition-colors ${
                isDone
                  ? 'bg-ok/5 border-ok/20 text-text-primary'
                  : isCurrent
                  ? 'bg-accent/5 border-accent/30 text-text-primary'
                  : 'bg-raised/20 border-separator/15 text-text-secondary/60'
              }`}
            >
              <div className="mt-0.5 shrink-0">
                {isDone ? (
                  <CheckCircle2 className="w-3.5 h-3.5 text-ok" />
                ) : isCurrent ? (
                  <Loader2 className="w-3.5 h-3.5 text-accent animate-spin" />
                ) : (
                  <Circle className="w-3.5 h-3.5 text-separator" />
                )}
              </div>
              <span className="text-[11px] leading-tight font-medium">
                {step.label}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
};
