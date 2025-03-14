import React from 'react';

// Helper function to format large numbers
export const formatLargeNumber = (value: string) => {
    if (!value) return 'N/A';
    // Truncate to 8 digits for display
    return value.substring(0, 8);
};

// Helper function to format percentiles for better display
export const formatPercentiles = (
    value: string | undefined
) => {
    if (!value) return 'N/A';

    // Split the percentile range
    const parts = value.split('~');
    if (parts.length !== 2) return value;

    return (
        <div className="flex flex-col items-center">
            <div className="flex items-center gap-2">
                <span>{parts[0]}%</span>
                <span className="text-[var(--muted-foreground)]">
                    ~
                </span>
                <span>{parts[1]}%</span>
            </div>
        </div>
    );
};

// Helper function to get color based on rank
export const getRankColor = (rank: string) => {
    if (!rank) return 'text-[var(--foreground)]';

    const rankNum = parseInt(rank);
    if (isNaN(rankNum)) return 'text-[var(--foreground)]';

    if (rankNum <= 10) return 'text-green-600 font-bold';
    if (rankNum <= 30) return 'text-blue-600';
    if (rankNum <= 50) return 'text-orange-600';
    return 'text-[var(--foreground)]';
};

// Helper function to get color based on score
export const getScoreColor = (score: number) => {
    if (score > 0.05) return 'text-green-600';
    if (score > 0) return 'text-blue-600';
    if (score > -0.05) return 'text-orange-600';
    return 'text-red-600';
};
