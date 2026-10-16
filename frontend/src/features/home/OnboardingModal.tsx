import React, { useState } from 'react';
import { Usb, KeyRound, ChevronRight, ChevronLeft, Info } from 'lucide-react';
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

export interface OnboardingModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

interface OnboardingStep {
  icon: React.ReactNode;
  titleKey: string;
  descKey: string;
  detailKey: string;
}

const STEPS: OnboardingStep[] = [
  {
    icon: <Usb className="w-5 h-5" />,
    titleKey: 'onboarding.step1Title',
    descKey: 'onboarding.step1Desc',
    detailKey: 'onboarding.step1Detail',
  },
  {
    icon: <KeyRound className="w-5 h-5" />,
    titleKey: 'onboarding.step2Title',
    descKey: 'onboarding.step2Desc',
    detailKey: 'onboarding.step2Detail',
  },
];

export const OnboardingModal: React.FC<OnboardingModalProps> = ({
  open,
  onOpenChange,
}) => {
  const { t } = useTranslation();
  const [currentStep, setCurrentStep] = useState(0);

  const step = STEPS[currentStep];
  const isLast = currentStep === STEPS.length - 1;
  const isFirst = currentStep === 0;

  const handleNext = () => {
    if (isLast) {
      onOpenChange(false);
      setCurrentStep(0);
    } else {
      setCurrentStep((s) => s + 1);
    }
  };

  const handleBack = () => {
    if (!isFirst) {
      setCurrentStep((s) => s - 1);
    }
  };

  const handleOpenChange = (nextOpen: boolean) => {
    if (!nextOpen) {
      setCurrentStep(0);
    }
    onOpenChange(nextOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t('onboarding.title')}</DialogTitle>
          <DialogDescription>{t('onboarding.description')}</DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          {/* Step indicator */}
          <div className="flex items-center gap-2">
            {STEPS.map((_, idx) => (
              <div
                key={idx}
                className={`h-1 flex-1 rounded-full transition-colors duration-200 ${
                  idx <= currentStep ? 'bg-accent' : 'bg-separator/40'
                }`}
              />
            ))}
          </div>

          <p className="text-[11px] text-text-secondary font-medium">
            {t('onboarding.stepNumber', {
              current: currentStep + 1,
              total: STEPS.length,
            })}
          </p>

          {/* Step content */}
          <div className="flex items-start gap-3 p-3 rounded-md bg-raised/50 border border-separator-subtle">
            <div className="w-9 h-9 rounded-md bg-accent/15 text-accent flex items-center justify-center shrink-0 mt-0.5">
              {step.icon}
            </div>
            <div className="space-y-1.5 min-w-0">
              <h3 className="text-sm font-semibold text-text-primary">
                {t(step.titleKey)}
              </h3>
              <p className="text-xs text-text-secondary leading-relaxed">
                {t(step.descKey)}
              </p>
            </div>
          </div>

          {/* Additional context */}
          <div className="flex items-start gap-2 px-1">
            <Info className="w-3.5 h-3.5 text-accent shrink-0 mt-0.5" />
            <p className="text-[11px] text-text-secondary leading-relaxed">
              {t(step.detailKey)}
            </p>
          </div>
        </div>

        <DialogFooter>
          {!isFirst && (
            <Button variant="ghost" size="sm" onClick={handleBack}>
              <ChevronLeft className="w-3.5 h-3.5" />
              {t('onboarding.back')}
            </Button>
          )}
          <Button variant="primary" size="sm" onClick={handleNext}>
            {isLast ? t('onboarding.done') : t('onboarding.next')}
            {!isLast && <ChevronRight className="w-3.5 h-3.5" />}
          </Button>
        </DialogFooter>

        <p className="text-center text-[11px] text-text-secondary/70 -mt-2">
          {t('onboarding.skipHint')}
        </p>
      </DialogContent>
    </Dialog>
  );
};
