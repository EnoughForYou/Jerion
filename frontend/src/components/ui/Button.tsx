import React from 'react';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'destructive' | 'ghost' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className = '',
      variant = 'secondary',
      size = 'md',
      isLoading = false,
      disabled,
      children,
      ...props
    },
    ref
  ) => {
    const baseStyles =
      'inline-flex items-center justify-center font-medium transition-all duration-150 cursor-pointer select-none disabled:opacity-50 disabled:pointer-events-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-1';

    const variantStyles = {
      primary: 'bg-accent text-white hover:opacity-90 active:scale-[0.99] border border-transparent shadow-sm',
      secondary: 'bg-raised text-text-primary border border-separator/40 hover:bg-surface/80 active:scale-[0.99]',
      destructive: 'bg-sev-critical text-white hover:opacity-90 active:scale-[0.99] border border-transparent shadow-sm',
      ghost: 'bg-transparent text-text-secondary hover:text-text-primary hover:bg-raised/40',
      outline: 'bg-transparent border border-separator text-text-primary hover:bg-raised/30',
    };

    const sizeStyles = {
      sm: 'h-7 px-2.5 text-xs rounded-sm gap-1.5',
      md: 'h-8 px-3.5 text-base rounded-md gap-2', // PRD §5.3 desktop default
      lg: 'h-10 px-5 text-md rounded-md gap-2.5',
    };

    return (
      <button
        ref={ref}
        disabled={disabled || isLoading}
        className={`${baseStyles} ${variantStyles[variant]} ${sizeStyles[size]} ${className}`}
        {...props}
      >
        {isLoading ? (
          <span className="inline-block h-3.5 w-3.5 border-2 border-current border-t-transparent rounded-full animate-spin" />
        ) : null}
        {children}
      </button>
    );
  }
);

Button.displayName = 'Button';
