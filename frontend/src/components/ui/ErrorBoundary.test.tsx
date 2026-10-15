import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React, { useState } from 'react';
import { createMemoryRouter, RouterProvider } from 'react-router-dom';
import { ErrorBoundary, RouteErrorBoundary } from './ErrorBoundary';
import '../../i18n';

// Component that throws on demand
const ProblemChild: React.FC<{ shouldThrow?: boolean }> = ({ shouldThrow = true }) => {
  if (shouldThrow) {
    throw new Error('Test crash in child component');
  }
  return <div>Safe child content</div>;
};

// Resettable error component
const FlakyComponent: React.FC = () => {
  const [fail, setFail] = useState(true);
  if (fail) {
    return (
      <div>
        <button onClick={() => setFail(false)}>Fix</button>
        <ProblemChild shouldThrow={true} />
      </div>
    );
  }
  return <div>Recovered Successfully</div>;
};

describe('ErrorBoundary', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    // Suppress expected React ErrorBoundary log output in test console
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <div>Hello World</div>
      </ErrorBoundary>
    );
    expect(screen.getByText('Hello World')).toBeInTheDocument();
  });

  it('catches render errors and displays fallback UI without crashing', () => {
    render(
      <ErrorBoundary>
        <ProblemChild />
      </ErrorBoundary>
    );

    expect(screen.getByText('Component Failed to Render')).toBeInTheDocument();
    expect(screen.getByText('Test crash in child component')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /try again/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /reload application/i })).toBeInTheDocument();
  });

  it('recovers when Try Again is clicked after fixing error condition', () => {
    render(
      <ErrorBoundary>
        <FlakyComponent />
      </ErrorBoundary>
    );

    expect(screen.getByText('Component Failed to Render')).toBeInTheDocument();
    const tryAgainBtn = screen.getByRole('button', { name: /try again/i });
    fireEvent.click(tryAgainBtn);
  });

  it('renders custom fallback when provided', () => {
    render(
      <ErrorBoundary fallback={<div>Custom Error Screen</div>}>
        <ProblemChild />
      </ErrorBoundary>
    );

    expect(screen.getByText('Custom Error Screen')).toBeInTheDocument();
    expect(screen.queryByText('Component Failed to Render')).not.toBeInTheDocument();
  });
});

describe('RouteErrorBoundary', () => {
  let consoleErrorSpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    consoleErrorSpy.mockRestore();
  });

  it('catches router errors and displays route recovery UI', () => {
    const testRouter = createMemoryRouter(
      [
        {
          path: '/',
          element: <ProblemChild shouldThrow={true} />,
          errorElement: <RouteErrorBoundary />,
        },
      ],
      { initialEntries: ['/'] }
    );

    render(<RouterProvider router={testRouter} />);

    expect(screen.getByText('View Recovery')).toBeInTheDocument();
    expect(screen.getByText('Test crash in child component')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /home/i })).toBeInTheDocument();
  });
});
