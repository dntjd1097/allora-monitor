import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { format, parseISO } from 'date-fns';

/**
 * Combines multiple class names with Tailwind CSS
 */
export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

/**
 * Formats a date string to a readable format
 */
export function formatDate(
    dateString: string,
    formatStr: string = 'PPP'
) {
    try {
        return format(parseISO(dateString), formatStr);
    } catch (_) {
        return dateString;
    }
}

/**
 * Formats a number with commas
 */
export function formatNumber(num: number): string {
    return new Intl.NumberFormat().format(num);
}

/**
 * Truncates a string to a specified length
 */
export function truncateString(
    str: string,
    length: number = 50
): string {
    if (!str) return '';
    if (str.length <= length) return str;
    return `${str.substring(0, length)}...`;
}

/**
 * Converts bytes to a human-readable format
 */
export function formatBytes(
    bytes: number,
    decimals: number = 2
): string {
    if (bytes === 0) return '0 Bytes';

    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = [
        'Bytes',
        'KB',
        'MB',
        'GB',
        'TB',
        'PB',
        'EB',
        'ZB',
        'YB',
    ];

    const i = Math.floor(Math.log(bytes) / Math.log(k));

    return (
        parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) +
        ' ' +
        sizes[i]
    );
}

/**
 * Generates a random color
 */
export function getRandomColor(): string {
    return `#${Math.floor(
        Math.random() * 16777215
    ).toString(16)}`;
}

/**
 * Debounces a function
 */
export function debounce<
    T extends (...args: unknown[]) => unknown
>(func: T, wait: number): (...args: Parameters<T>) => void {
    let timeout: NodeJS.Timeout;

    return function (...args: Parameters<T>): void {
        clearTimeout(timeout);
        timeout = setTimeout(() => func(...args), wait);
    };
}
