import React from 'react';
import { clsx } from 'clsx';
import type { BaseComponentProps } from '../../types/common';

interface CardProps extends BaseComponentProps {
  padding?: 'none' | 'sm' | 'md' | 'lg';
  shadow?: 'none' | 'sm' | 'md' | 'lg';
  hover?: boolean;
}

const Card: React.FC<CardProps> = ({
  children,
  className,
  padding = 'md',
  shadow = 'sm',
  hover = false,
  ...props
}) => {
  const baseClasses = 'bg-white rounded-lg border border-gray-200';
  
  const paddingClasses = {
    none: '',
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
  };
  
  const shadowClasses = {
    none: '',
    sm: 'shadow-sm',
    md: 'shadow-md',
    lg: 'shadow-lg',
  };
  
  const hoverClasses = hover ? 'transition-shadow hover:shadow-md' : '';
  
  const classes = clsx(
    baseClasses,
    paddingClasses[padding],
    shadowClasses[shadow],
    hoverClasses,
    className
  );

  return (
    <div className={classes} {...props}>
      {children}
    </div>
  );
};

// Card sub-components
export const CardHeader: React.FC<BaseComponentProps> = ({ children, className, ...props }) => (
  <div className={clsx('border-b border-gray-200 pb-4 mb-4', className)} {...props}>
    {children}
  </div>
);

export const CardBody: React.FC<BaseComponentProps> = ({ children, className, ...props }) => (
  <div className={clsx('', className)} {...props}>
    {children}
  </div>
);

export const CardFooter: React.FC<BaseComponentProps> = ({ children, className, ...props }) => (
  <div className={clsx('border-t border-gray-200 pt-4 mt-4', className)} {...props}>
    {children}
  </div>
);

export const CardTitle: React.FC<BaseComponentProps> = ({ children, className, ...props }) => (
  <h3 className={clsx('text-lg font-semibold text-gray-900', className)} {...props}>
    {children}
  </h3>
);

export const CardDescription: React.FC<BaseComponentProps> = ({ children, className, ...props }) => (
  <p className={clsx('text-sm text-gray-600 mt-1', className)} {...props}>
    {children}
  </p>
);

export default Card;