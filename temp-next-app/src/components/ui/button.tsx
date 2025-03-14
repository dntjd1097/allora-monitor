import React from 'react';
import { cn } from '@/lib/utils';

interface ButtonProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'default' | 'outline' | 'ghost' | 'link';
    size?: 'default' | 'sm' | 'lg' | 'icon';
}

export function Button({
    className,
    variant = 'default',
    size = 'default',
    children,
    ...props
}: ButtonProps) {
    return (
        <button
            className={cn(
                'inline-flex items-center justify-center rounded-md font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none',
                {
                    'bg-primary text-primary-foreground hover:bg-primary/90':
                        variant === 'default',
                    'border border-input hover:bg-accent hover:text-accent-foreground':
                        variant === 'outline',
                    'hover:bg-accent hover:text-accent-foreground':
                        variant === 'ghost',
                    'underline-offset-4 hover:underline text-primary':
                        variant === 'link',
                },
                {
                    'h-10 py-2 px-4': size === 'default',
                    'h-9 px-3 rounded-md': size === 'sm',
                    'h-11 px-8 rounded-md': size === 'lg',
                    'h-10 w-10': size === 'icon',
                },
                className
            )}
            {...props}
        >
            {children}
        </button>
    );
}
