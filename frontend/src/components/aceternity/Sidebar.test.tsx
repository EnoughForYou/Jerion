import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import { Shield } from 'lucide-react';
import {
  SidebarProvider,
  Sidebar,
  SidebarBody,
  SidebarLink,
  SidebarToggle,
} from './Sidebar';

describe('Sidebar component', () => {
  it('renders expanded by default with label visible', () => {
    render(
      <SidebarProvider defaultOpen={true}>
        <Sidebar>
          <SidebarBody>
            <SidebarLink icon={<Shield className="w-4 h-4" />} label="Overview" active />
            <SidebarToggle />
          </SidebarBody>
        </Sidebar>
      </SidebarProvider>
    );

    const aside = screen.getByRole('navigation');
    expect(aside).toHaveClass('w-[220px]');
    expect(screen.getByText('Overview')).toBeInTheDocument();
  });

  it('collapses to 64px when toggle button is clicked', () => {
    render(
      <SidebarProvider defaultOpen={true}>
        <Sidebar>
          <SidebarBody>
            <SidebarLink icon={<Shield className="w-4 h-4" />} label="Overview" />
            <SidebarToggle />
          </SidebarBody>
        </Sidebar>
      </SidebarProvider>
    );

    const aside = screen.getByRole('navigation');
    const toggleBtn = screen.getByRole('button', { name: /collapse sidebar/i });
    fireEvent.click(toggleBtn);

    expect(aside).toHaveClass('w-[64px]');
    expect(screen.queryByText('Overview')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /expand sidebar/i })).toBeInTheDocument();
  });

  it('indicates active page via aria-current', () => {
    render(
      <SidebarProvider defaultOpen={true}>
        <Sidebar>
          <SidebarBody>
            <SidebarLink icon={<Shield className="w-4 h-4" />} label="Active Tab" active />
            <SidebarLink icon={<Shield className="w-4 h-4" />} label="Inactive Tab" active={false} />
          </SidebarBody>
        </Sidebar>
      </SidebarProvider>
    );

    const activeBtn = screen.getByRole('button', { name: 'Active Tab' });
    expect(activeBtn).toHaveAttribute('aria-current', 'page');

    const inactiveBtn = screen.getByRole('button', { name: 'Inactive Tab' });
    expect(inactiveBtn).not.toHaveAttribute('aria-current');
  });
});
