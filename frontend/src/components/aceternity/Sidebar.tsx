import React, { useState } from 'react';
import { PanelLeftClose, PanelLeftOpen } from 'lucide-react';
import { SimpleTooltip } from '../ui/Tooltip';
import { SidebarContext, useSidebar } from './SidebarContext';

export interface SidebarProviderProps {
  children: React.ReactNode;
  defaultOpen?: boolean;
}

export const SidebarProvider: React.FC<SidebarProviderProps> = ({
  children,
  defaultOpen = true,
}) => {
  const [open, setOpen] = useState(defaultOpen);

  const toggle = () => setOpen((prev) => !prev);

  return (
    <SidebarContext.Provider value={{ open, setOpen, toggle }}>
      {children}
    </SidebarContext.Provider>
  );
};

export interface SidebarProps extends React.HTMLAttributes<HTMLElement> {
  children: React.ReactNode;
}

export const Sidebar = React.forwardRef<HTMLElement, SidebarProps>(
  ({ children, className = '', ...props }, ref) => {
    const { open } = useSidebar();

    return (
      <aside
        ref={ref}
        role="navigation"
        aria-label="Main Navigation"
        className={`h-full shrink-0 border-r border-separator/35 bg-surface/90 backdrop-blur-md flex flex-col justify-between transition-[width] duration-200 ease-out select-none z-20 ${
          open ? 'w-[220px]' : 'w-[64px]'
        } ${className}`}
        {...props}
      >
        {children}
      </aside>
    );
  }
);
Sidebar.displayName = 'Sidebar';

export const SidebarBody: React.FC<{ children: React.ReactNode; className?: string }> = ({
  children,
  className = '',
}) => {
  return (
    <div className={`flex flex-col h-full justify-between p-2.5 ${className}`}>
      {children}
    </div>
  );
};

export interface SidebarLinkProps {
  icon: React.ReactNode;
  label: string;
  active?: boolean;
  badge?: React.ReactNode;
  onClick?: () => void;
  className?: string;
}

export const SidebarLink: React.FC<SidebarLinkProps> = ({
  icon,
  label,
  active = false,
  badge,
  onClick,
  className = '',
}) => {
  const { open } = useSidebar();

  const content = (
    <button
      type="button"
      onClick={onClick}
      aria-label={label}
      aria-current={active ? 'page' : undefined}
      className={`group flex items-center w-full h-9 rounded-md text-xs font-medium transition-colors duration-150 cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
        open ? 'px-2.5 gap-2.5 justify-start' : 'px-0 justify-center'
      } ${
        active
          ? 'bg-raised text-text-primary shadow-xs'
          : 'text-text-secondary hover:text-text-primary hover:bg-raised/40'
      } ${className}`}
    >
      <span
        className={`shrink-0 transition-colors ${
          active ? 'text-accent' : 'text-text-secondary group-hover:text-text-primary'
        }`}
      >
        {icon}
      </span>
      {open ? (
        <span className="truncate flex-1 text-left">{label}</span>
      ) : null}
      {open && badge ? <span className="shrink-0">{badge}</span> : null}
    </button>
  );

  if (!open) {
    return (
      <SimpleTooltip content={label} side="right">
        {content}
      </SimpleTooltip>
    );
  }

  return content;
};

export const SidebarToggle: React.FC<{ className?: string }> = ({ className = '' }) => {
  const { open, toggle } = useSidebar();

  return (
    <button
      type="button"
      onClick={toggle}
      aria-label={open ? 'Collapse Sidebar' : 'Expand Sidebar'}
      className={`flex items-center justify-center h-8 rounded-md text-text-secondary hover:text-text-primary hover:bg-raised/40 transition-colors cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent ${
        open ? 'w-full gap-2 px-2.5 text-xs' : 'w-full px-0'
      } ${className}`}
    >
      {open ? (
        <>
          <PanelLeftClose className="w-4 h-4 shrink-0" />
          <span className="truncate flex-1 text-left">Collapse</span>
        </>
      ) : (
        <PanelLeftOpen className="w-4 h-4 shrink-0" />
      )}
    </button>
  );
};
