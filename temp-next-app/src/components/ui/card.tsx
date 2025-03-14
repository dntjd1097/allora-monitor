import React from 'react';
import { cn } from '@/lib/utils';

interface CardProps
    extends React.HTMLAttributes<HTMLDivElement> {
    title?: string;
    description?: string;
}

export function Card({
    className,
    title,
    description,
    children,
    ...props
}: CardProps) {
    return (
        <div
            className={cn(
                'card overflow-hidden',
                className
            )}
            {...props}
        >
            {(title || description) && (
                <div className="mb-4">
                    {title && (
                        <h3 className="text-lg font-semibold">
                            {title}
                        </h3>
                    )}
                    {description && (
                        <p className="text-sm text-[var(--muted-foreground)]">
                            {description}
                        </p>
                    )}
                </div>
            )}
            {children}
        </div>
    );
}

interface CardHeaderProps
    extends React.HTMLAttributes<HTMLDivElement> {}

export function CardHeader({
    className,
    children,
    ...props
}: CardHeaderProps) {
    return (
        <div
            className={cn(
                'flex flex-col space-y-1.5 pb-4',
                className
            )}
            {...props}
        >
            {children}
        </div>
    );
}

interface CardTitleProps
    extends React.HTMLAttributes<HTMLHeadingElement> {}

export function CardTitle({
    className,
    children,
    ...props
}: CardTitleProps) {
    return (
        <h3
            className={cn(
                'text-lg font-semibold leading-none tracking-tight',
                className
            )}
            {...props}
        >
            {children}
        </h3>
    );
}

interface CardContentProps
    extends React.HTMLAttributes<HTMLDivElement> {}

export function CardContent({
    className,
    children,
    ...props
}: CardContentProps) {
    return (
        <div className={cn('pt-0', className)} {...props}>
            {children}
        </div>
    );
}

interface CardFooterProps
    extends React.HTMLAttributes<HTMLDivElement> {}

export function CardFooter({
    className,
    children,
    ...props
}: CardFooterProps) {
    return (
        <div
            className={cn(
                'flex items-center pt-4',
                className
            )}
            {...props}
        >
            {children}
        </div>
    );
}
