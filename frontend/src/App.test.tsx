import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import App from './App';

vi.mock('../wailsjs/go/main/App', () => ({
  Greet: vi.fn().mockResolvedValue('Hello Jerion'),
}));

describe('App Component', () => {
  it('renders initial prompt and greet button', () => {
    render(<App />);
    expect(screen.getByText(/Please enter your name below/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Greet/i })).toBeInTheDocument();
  });
});
