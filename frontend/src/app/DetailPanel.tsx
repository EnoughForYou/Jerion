import React from 'react';
import { X } from 'lucide-react';
import { Button } from '../components/ui/Button';

export interface DetailPanelProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  children: React.ReactNode;
}

export const DetailPanel: React.FC<DetailPanelProps> = ({
  open,
  onClose,
  title,
  description,
  children,
}) => {
  if (!open) return null;

  return (
    <aside
      role="complementary"
      aria-label={title}
      className="w-[420px] h-full shrink-0 border-l border-separator-subtle bg-surface/95 backdrop-blur-md flex flex-col justify-between shadow-xl z-20 select-none animate-in slide-in-from-right duration-200"
    >
      {/* Panel Header */}
      <div className="h-12 px-4 border-b border-separator-muted flex items-center justify-between shrink-0">
        <div className="flex flex-col truncate pr-2">
          <h3 className="text-base font-semibold text-text-primary truncate">{title}</h3>
          {description ? (
            <p className="text-[11px] text-text-secondary truncate">{description}</p>
          ) : null}
        </div>
        <Button variant="ghost" size="sm" onClick={onClose} aria-label="Close detail panel">
          <X className="w-4 h-4" />
        </Button>
      </div>

      {/* Panel Content */}
      <div className="flex-1 overflow-y-auto p-4 select-text">
        {children}
      </div>
    </aside>
  );
};
