import React from 'react';
import { cn } from '@/lib/utils';

interface BadgeProps
    extends React.HTMLAttributes<HTMLSpanElement> {
    variant?:
        | 'default'
        | 'success'
        | 'warning'
        | 'error'
        | 'info';
}

export function Badge({
    className,
    variant = 'default',
    children,
    ...props
}: BadgeProps) {
    const variantClasses = {
        default:
            'bg-[var(--secondary)] text-[var(--secondary-foreground)]',
        success:
            'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300',
        warning:
            'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
        error: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
        info: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
    };

    return (
        <span
            className={cn(
                'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
                variantClasses[variant],
                className
            )}
            {...props}
        >
            {children}
        </span>
    );
}
