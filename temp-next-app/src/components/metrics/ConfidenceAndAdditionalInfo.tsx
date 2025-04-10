import React from 'react';
import { TopicInferenceData } from '@/lib/api';
import { formatDate } from '@/lib/utils';
import { formatLargeNumber } from '@/lib/formatters';

interface ConfidenceAndAdditionalInfoProps {
    inferenceData: TopicInferenceData | null;
}

export const ConfidenceAndAdditionalInfo: React.FC<
    ConfidenceAndAdditionalInfoProps
> = ({ inferenceData }) => {
    if (!inferenceData || !inferenceData.data) {
        return (
            <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                <p className="text-yellow-800">
                    No inference data available
                </p>
            </div>
        );
    }

    const {
        confidence_interval_raw_percentiles,
        confidence_interval_values,
        timestamp,
    } = inferenceData.data;

    if (
        !confidence_interval_raw_percentiles ||
        !confidence_interval_values
    ) {
        return (
            <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
                <p className="text-yellow-800">
                    Confidence interval data is not
                    available
                </p>
            </div>
        );
    }

    return (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            <div>
                <h2 className="text-xl font-semibold mb-4 flex items-center text-[var(--foreground)] border-b pb-2 border-gray-200">
                    <span className="mr-2">
                        Confidence Intervals
                    </span>
                    <span className="text-sm text-[var(--muted-foreground)] font-normal">
                        (Percentiles and their values)
                    </span>
                </h2>
                <div className="overflow-x-auto bg-white rounded-lg shadow-sm border border-gray-200">
                    <table className="w-full border-collapse">
                        <thead>
                            <tr className="bg-gray-100">
                                <th className="border border-gray-300 p-2 text-left font-semibold text-[var(--foreground)]">
                                    Percentile
                                </th>
                                <th className="border border-gray-300 p-2 text-left font-semibold text-[var(--foreground)]">
                                    Value
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {confidence_interval_raw_percentiles.map(
                                (percentile, index) => (
                                    <tr
                                        key={percentile}
                                        className={
                                            index % 2 === 0
                                                ? 'bg-gray-50'
                                                : ''
                                        }
                                    >
                                        <td className="border border-gray-300 p-2">
                                            <span className="font-medium">
                                                {percentile}
                                                %
                                            </span>
                                        </td>
                                        <td className="border border-gray-300 p-2 font-mono">
                                            {formatLargeNumber(
                                                confidence_interval_values[
                                                    index
                                                ]
                                            )}
                                        </td>
                                    </tr>
                                )
                            )}
                        </tbody>
                    </table>
                </div>
            </div>

            <div>
                <h2 className="text-xl font-semibold mb-4 text-[var(--foreground)] flex items-center border-b pb-2 border-gray-200">
                    <span className="mr-2">
                        Additional Information
                    </span>
                    <span className="text-sm text-[var(--muted-foreground)] font-normal">
                        (Metadata)
                    </span>
                </h2>
                <div className="grid grid-cols-1 gap-4">
                    <div className="border border-gray-300 rounded-lg shadow-sm p-4 bg-white hover:shadow-md transition-shadow">
                        <div>
                            <h3 className="font-medium mb-2 text-[var(--foreground)]">
                                Timestamp
                            </h3>
                            <p className="text-sm bg-gray-50 p-3 rounded font-mono">
                                {formatDate(
                                    timestamp,
                                    'PPpp'
                                )}
                            </p>
                            <p className="text-xs text-[var(--muted-foreground)] mt-1">
                                (UTC:{' '}
                                {new Date(
                                    timestamp
                                ).toISOString()}
                                )
                            </p>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
};
