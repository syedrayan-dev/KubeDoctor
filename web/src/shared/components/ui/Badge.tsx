import React from 'react';
import { clsx } from 'clsx';
import { STATUS_COLORS } from '../../constants/colors';
import type { BaseComponentProps, RequestStatus, Size } from '../../types/common';

interface BadgeProps extends BaseComponentProps {
  variant?: 'default' | 'outline' | 'status' | 'success' | 'warning' | 'error' | 'info';
  size?: Size;
  status?: RequestStatus;
}

const Badge: React.FC<BadgeProps> = ({
  children,
  className,
  variant = 'default',
  size = 'md',
  status,
  ...props
}) => {
  const baseClasses = 'inline-flex items-center font-medium rounded-full border';
  
  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-0.5 text-xs',
    lg: 'px-3 py-1 text-sm',
    xl: 'px-4 py-1.5 text-sm',
  };
  
  const variantClasses = {
    default: 'bg-white text-gray-600 border-gray-300',
    outline: 'bg-white text-gray-700 border-gray-300',
    success: 'bg-white text-green-600 border-green-300',
    warning: 'bg-white text-yellow-600 border-yellow-300',
    error: 'bg-white text-red-600 border-red-300',
    info: 'bg-white text-blue-600 border-blue-300',
    status: '', // Will be overridden by status colors
  };
  
  // Use status-specific colors if status is provided
  const statusClasses = status && variant === 'status' ? STATUS_COLORS[status] : '';
  
  const classes = clsx(
    baseClasses,
    sizeClasses[size],
    statusClasses || variantClasses[variant],
    className
  );

  return (
    <span className={classes} {...props}>
      {children}
    </span>
  );
};

export default Badge;