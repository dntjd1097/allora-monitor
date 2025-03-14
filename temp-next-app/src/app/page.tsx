'use client';

import {
    useEffect,
    useState,
    useRef,
    useCallback,
} from 'react';
import { Loading } from '@/components/ui/loading';
import {
    Competition,
    fetchActiveCompetitions,
    fetchTopicInference,
    fetchTopicHeights,
    TopicInferenceData,
} from '@/lib/api';
import { CompetitionSelector } from '@/components/competition/CompetitionSelector';
import { MetricsDisplay } from '@/components/metrics/MetricsDisplay';
import { PaginationControls } from '@/components/pagination/PaginationControls';
import { SynthesisTable } from '@/components/table/SynthesisTable';
import { ConfidenceAndAdditionalInfo } from '@/components/metrics/ConfidenceAndAdditionalInfo';
import { SynthesisValueItem } from '@/types';

export default function HomePage() {
    const [competitions, setCompetitions] = useState<
        Competition[]
    >([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedCompetition, setSelectedCompetition] =
        useState<Competition | null>(null);
    const [inferenceData, setInferenceData] =
        useState<TopicInferenceData | null>(null);
    const [inferenceLoading, setInferenceLoading] =
        useState<boolean>(false);

    // Pagination state
    const [blockHeight, setBlockHeight] = useState<
        string | null
    >(null);
    const [availableHeights, setAvailableHeights] =
        useState<string[]>([]);
    const [heightsLoading, setHeightsLoading] =
        useState<boolean>(false);
    const [customHeight, setCustomHeight] =
        useState<string>('');
    const [showAllHeights, setShowAllHeights] =
        useState<boolean>(false);

    // Auto refresh state - 기본값을 true로 설정
    const [autoRefreshHeights] = useState<boolean>(true);
    const intervalRef = useRef<NodeJS.Timeout | null>(null);

    // Style for better text readability on mobile
    const textStyle = {
        color: '#000000',
        fontWeight: 600,
    };

    // Function to fetch inference data
    const fetchInferenceData = useCallback(async () => {
        if (!selectedCompetition) return;

        try {
            setInferenceLoading(true);
            // Use the blockHeight parameter directly with the API function
            const data = await fetchTopicInference(
                selectedCompetition.topic_id.toString(),
                blockHeight || undefined
            );
            console.log('Fetched inference data:', data);
            setInferenceData(data);
        } catch (err) {
            console.error(
                'Error fetching inference data:',
                err
            );
        } finally {
            setInferenceLoading(false);
        }
    }, [selectedCompetition, blockHeight]);

    // Function to fetch heights
    const fetchHeights = useCallback(async () => {
        if (!selectedCompetition) return;

        try {
            setHeightsLoading(true);
            const response = await fetchTopicHeights(
                selectedCompetition.topic_id.toString()
            );
            if (response.data && response.data.heights) {
                setAvailableHeights(response.data.heights);
            }
        } catch (err) {
            console.error('Error fetching heights:', err);
        } finally {
            setHeightsLoading(false);
        }
    }, [selectedCompetition]);

    // Toggle show all heights
    const toggleShowAllHeights = () => {
        setShowAllHeights(!showAllHeights);
    };

    // Handle competition selection
    const handleSelectCompetition = (
        competition: Competition
    ) => {
        setSelectedCompetition(competition);
        setBlockHeight(null);
    };

    // Extract synthesis values from inference data
    const getSynthesisValues = (): SynthesisValueItem[] => {
        if (
            !inferenceData?.data?.network_inferences
                ?.synthesis_value
        ) {
            return [];
        }
        return inferenceData.data.network_inferences
            .synthesis_value;
    };

    // Initial fetch of competitions
    useEffect(() => {
        const fetchCompetitions = async () => {
            try {
                setLoading(true);
                const data =
                    await fetchActiveCompetitions();
                console.log('Fetched competitions:', data);
                setCompetitions(data);
                if (data.length > 0) {
                    setSelectedCompetition(data[0]);
                }
            } catch (err) {
                console.error(
                    'Error fetching competitions:',
                    err
                );
                setError('Failed to fetch competitions');
            } finally {
                setLoading(false);
            }
        };

        fetchCompetitions();
    }, []);

    // Effect for fetching inference data when competition or block height changes
    useEffect(() => {
        fetchInferenceData();
    }, [fetchInferenceData]);

    // Effect for fetching heights when competition changes
    useEffect(() => {
        fetchHeights();
    }, [fetchHeights]);

    // Effect for auto refresh of heights only
    useEffect(() => {
        // Clear any existing interval
        if (intervalRef.current) {
            clearInterval(intervalRef.current);
            intervalRef.current = null;
        }

        // Set up new interval if auto refresh is enabled
        if (autoRefreshHeights) {
            intervalRef.current = setInterval(() => {
                console.log('Auto refreshing heights...');
                fetchHeights();
            }, 60000); // 1 minute
        }

        // Cleanup on unmount or when autoRefreshHeights changes
        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
                intervalRef.current = null;
            }
        };
    }, [autoRefreshHeights, fetchHeights]);

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <Loading
                    size="lg"
                    text="Loading competitions..."
                />
            </div>
        );
    }

    if (error) {
        return (
            <div className="flex flex-col items-center justify-center p-8">
                <div className="text-red-500 mb-4">
                    ⚠️ {error}
                </div>
            </div>
        );
    }

    return (
        <div
            className="bg-gradient-to-b from-gray-50 to-gray-100"
            style={textStyle}
        >
            <div className="border border-gray-300 rounded-lg overflow-hidden shadow-lg bg-white">
                {/* Competition Selector Component */}
                <CompetitionSelector
                    competitions={competitions}
                    selectedCompetition={
                        selectedCompetition
                    }
                    onSelectCompetition={
                        handleSelectCompetition
                    }
                />

                <div className="p-6">
                    {inferenceLoading ? (
                        <div className="flex justify-center items-center h-40">
                            <Loading
                                size="md"
                                text="Loading inference data..."
                            />
                        </div>
                    ) : inferenceData ? (
                        <div>
                            {/* Metrics Display Component */}
                            <MetricsDisplay
                                inferenceData={
                                    inferenceData
                                }
                            />

                            {/* Pagination Controls Component */}
                            {inferenceData.pagination && (
                                <PaginationControls
                                    pagination={
                                        inferenceData.pagination
                                    }
                                    blockHeight={
                                        blockHeight
                                    }
                                    setBlockHeight={
                                        setBlockHeight
                                    }
                                    availableHeights={
                                        availableHeights
                                    }
                                    heightsLoading={
                                        heightsLoading
                                    }
                                    customHeight={
                                        customHeight
                                    }
                                    setCustomHeight={
                                        setCustomHeight
                                    }
                                    showAllHeights={
                                        showAllHeights
                                    }
                                    toggleShowAllHeights={
                                        toggleShowAllHeights
                                    }
                                />
                            )}

                            {/* Synthesis Table Component */}
                            <SynthesisTable
                                synthesisValues={getSynthesisValues()}
                            />

                            {/* Confidence Intervals and Additional Info Component */}
                            <ConfidenceAndAdditionalInfo
                                inferenceData={
                                    inferenceData
                                }
                            />
                        </div>
                    ) : (
                        <div
                            className="text-center text-[var(--muted-foreground)] p-8 bg-gray-50 rounded-lg"
                            style={textStyle}
                        >
                            <div className="text-4xl mb-2">
                                📊
                            </div>
                            No inference data available
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
