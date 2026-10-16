import React, { Component, type ErrorInfo, type ReactNode } from 'react';
import { useRouteError, useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { RotateCcw, Home, AlertOctagon } from 'lucide-react';
import { Button } from './Button';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onReset?: () => void;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  public state: ErrorBoundaryState = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('ErrorBoundary caught an unhandled error:', error, errorInfo);
  }

  public reset = (): void => {
    this.props.onReset?.();
    this.setState({ hasError: false, error: null });
  };

  public render(): ReactNode {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="flex flex-col items-center justify-center p-6 text-center border border-sev-critical/30 rounded-lg bg-sev-critical/5 max-w-lg mx-auto my-8">
          <div className="w-12 h-12 rounded-full bg-sev-critical/15 text-sev-critical flex items-center justify-center mb-3">
            <AlertOctagon className="w-6 h-6" />
          </div>
          <span className="text-[10px] font-mono uppercase tracking-wider text-sev-critical mb-1 font-semibold">
            Application Error
          </span>
          <h3 className="text-base font-semibold text-text-primary mb-1">
            Component Failed to Render
          </h3>
          <p className="text-xs text-text-secondary leading-relaxed mb-4 max-w-sm">
            {this.state.error?.message || 'An unexpected runtime error occurred.'}
          </p>
          <div className="flex items-center gap-2.5">
            <Button variant="secondary" size="sm" onClick={this.reset}>
              <RotateCcw className="w-3.5 h-3.5" />
              Try Again
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                window.location.hash = '#/';
                window.location.reload();
              }}
            >
              <Home className="w-3.5 h-3.5" />
              Reload Application
            </Button>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

/**
 * RouteErrorBoundary handles routing level errors thrown in React Router loaders,
 * render cycles, or route components.
 */
export const RouteErrorBoundary: React.FC = () => {
  const error = useRouteError();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const errorMessage =
    error instanceof Error
      ? error.message
      : typeof error === 'string'
      ? error
      : t('common.errorRemedy', 'An unexpected error occurred while loading this view.');

  return (
    <div className="flex flex-col items-center justify-center min-h-[360px] p-8 text-center border border-separator-subtle rounded-lg bg-surface max-w-lg mx-auto my-12 shadow-sm">
      <div className="w-12 h-12 rounded-full bg-sev-critical/15 text-sev-critical flex items-center justify-center mb-4">
        <AlertOctagon className="w-6 h-6" />
      </div>
      <span className="text-[10px] font-mono uppercase tracking-wider text-sev-critical mb-1 font-semibold">
        View Recovery
      </span>
      <h3 className="text-base font-semibold text-text-primary mb-2">
        {t('common.error', 'Failed to Load View')}
      </h3>
      <p className="text-xs text-text-secondary leading-relaxed mb-6 max-w-md font-mono bg-raised/50 p-2.5 rounded border border-separator-muted break-words">
        {errorMessage}
      </p>
      <div className="flex items-center gap-3">
        <Button
          variant="secondary"
          size="sm"
          onClick={() => {
            navigate(0); // refresh current route
          }}
        >
          <RotateCcw className="w-3.5 h-3.5" />
          <span>{t('device.refresh', 'Refresh')}</span>
        </Button>
        <Button
          variant="primary"
          size="sm"
          onClick={() => {
            navigate('/');
          }}
        >
          <Home className="w-3.5 h-3.5" />
          <span>{t('nav.home', 'Home')}</span>
        </Button>
      </div>
    </div>
  );
};
