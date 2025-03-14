import React from 'react';
import { cn } from '@/lib/utils';

interface LoadingProps
    extends React.HTMLAttributes<HTMLDivElement> {
    size?: 'sm' | 'md' | 'lg';
    text?: string;
}

// 모바일 환경에서 텍스트 가독성 향상을 위한 스타일
const textStyle = {
    color: '#000000',
    fontWeight: 600,
};

export function Loading({
    className,
    size = 'md',
    text,
    ...props
}: LoadingProps) {
    const sizeClasses = {
        sm: 'w-4 h-4 border-2',
        md: 'w-8 h-8 border-3',
        lg: 'w-12 h-12 border-4',
    };

    return (
        <div
            className={cn(
                'flex flex-col items-center justify-center p-4',
                className
            )}
            {...props}
        >
            <div
                className={cn(
                    'animate-spin rounded-full border-solid border-[var(--primary)] border-t-transparent',
                    sizeClasses[size]
                )}
            />
            {text && (
                <p
                    className="mt-2 text-sm text-[var(--muted-foreground)]"
                    style={textStyle}
                >
                    {text}
                </p>
            )}
        </div>
    );
}
