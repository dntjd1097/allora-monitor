import React from 'react';
import { TopicInferenceData } from '@/lib/api';
import { formatLargeNumber } from '@/lib/formatters';

interface MetricsDisplayProps {
    inferenceData: TopicInferenceData;
}

export const MetricsDisplay: React.FC<
    MetricsDisplayProps
> = ({ inferenceData }) => {
    return (
        <>
            <h2
                className="text-xl font-semibold mb-4 text-[var(--foreground)] border-b pb-2 border-gray-200"
                style={{
                    color: '#000000',
                    fontWeight: 600,
                }}
            >
                Metrics
            </h2>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
                <div className="border border-gray-300 rounded-lg shadow-sm hover:shadow-md transition-shadow p-4 bg-white">
                    <div>
                        <h3 className="text-center font-medium text-[var(--foreground)] mb-2">
                            Combined Value
                        </h3>
                        <p className="text-2xl font-bold text-center text-blue-600">
                            {formatLargeNumber(
                                inferenceData.data
                                    .network_inferences
                                    .combined_value
                            )}
                        </p>
                    </div>
                </div>

                <div className="border border-gray-300 rounded-lg shadow-sm hover:shadow-md transition-shadow p-4 bg-white">
                    <div>
                        <h3 className="text-center font-medium text-[var(--foreground)] mb-2">
                            Naive Value
                        </h3>
                        <p className="text-2xl font-bold text-center text-green-600">
                            {formatLargeNumber(
                                inferenceData.data
                                    .network_inferences
                                    .naive_value
                            )}
                        </p>
                    </div>
                </div>

                <div className="border border-gray-300 rounded-lg shadow-sm hover:shadow-md transition-shadow p-4 bg-white">
                    <div>
                        <h3 className="text-center font-medium text-[var(--foreground)] mb-2">
                            Inference Block Height
                        </h3>
                        <p className="text-2xl font-bold text-center text-purple-600">
                            {
                                inferenceData.data
                                    .inference_block_height
                            }
                        </p>
                    </div>
                </div>

                <div className="border border-gray-300 rounded-lg shadow-sm hover:shadow-md transition-shadow p-4 bg-white">
                    <div>
                        <h3 className="text-center font-medium text-[var(--foreground)] mb-2">
                            Loss Block Height
                        </h3>
                        <p className="text-2xl font-bold text-center text-orange-600">
                            {
                                inferenceData.data
                                    .loss_block_height
                            }
                        </p>
                    </div>
                </div>
            </div>
        </>
    );
};
