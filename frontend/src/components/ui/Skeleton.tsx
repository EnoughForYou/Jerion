import React from 'react';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'rectangular' | 'circular' | 'text';
}

export const Skeleton: React.FC<SkeletonProps> = ({
  className = '',
  variant = 'rectangular',
  ...props
}) => {
  const variantStyles = {
    rectangular: 'rounded-md',
    circular: 'rounded-full',
    text: 'rounded-sm h-3 w-full my-1',
  };

  return (
    <div
      className={`animate-pulse bg-raised/60 ${variantStyles[variant]} ${className}`}
      {...props}
    />
  );
};
