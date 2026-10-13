import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import App from './App';

describe('App Component', () => {
  it('renders preflight on launch', async () => {
    render(<App />);
    expect(screen.getByText('System Preflight')).toBeInTheDocument();
    await waitFor(() => {
      // Mock preflight passes and transitions to main shell home view
      expect(screen.getByText('Device Overview')).toBeInTheDocument();
    });
  });
});
