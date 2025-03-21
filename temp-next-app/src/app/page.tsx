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
    fetchInactiveCompetitions,
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
    const [activeCompetitions, setActiveCompetitions] =
        useState<Competition[]>([]);
    const [inactiveCompetitions, setInactiveCompetitions] =
        useState<Competition[]>([]);
    const [loading, setLoading] = useState<boolean>(true);
    const [error, setError] = useState<string | null>(null);
    const [selectedCompetition, setSelectedCompetition] =
        useState<Competition | null>(null);
    const [inferenceData, setInferenceData] =
        useState<TopicInferenceData | null>(null);
    const [inferenceLoading, setInferenceLoading] =
        useState<boolean>(false);
    const [inferenceError, setInferenceError] = useState<
        string | null
    >(null);

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

    // Add a new state for checking if the competition has data
    const [hasNoData, setHasNoData] =
        useState<boolean>(false);

    // Style for better text readability on mobile
    const textStyle = {
        color: '#000000',
        fontWeight: 600,
    };

    // Add a ref to track the current request
    const currentRequestIdRef = useRef<number>(0);

    // Function to fetch inference data
    const fetchInferenceData = useCallback(async () => {
        if (!selectedCompetition) return;

        // Generate a unique request ID to track this specific request
        const requestId = ++currentRequestIdRef.current;

        try {
            setInferenceLoading(true);
            setInferenceError(null); // Reset error state
            setHasNoData(false); // Reset no data state

            // Use the blockHeight parameter directly with the API function
            const data = await fetchTopicInference(
                selectedCompetition.topic_id.toString(),
                blockHeight || undefined
            );

            // Check if this is still the most recent request
            if (requestId !== currentRequestIdRef.current) {
                console.log(
                    'Ignoring stale response for topic_id:',
                    selectedCompetition.topic_id
                );
                return; // Ignore stale responses
            }

            console.log('Fetched inference data:', data);

            if (data === null) {
                // Handle the case when no data is available
                setInferenceError(
                    `No inference data available for ${selectedCompetition.name} (Topic ID: ${selectedCompetition.topic_id})`
                );
                setInferenceData(null);
                setHasNoData(true);
            } else {
                setInferenceData(data);
            }
        } catch (err) {
            // Only update state if this is still the most recent request
            if (requestId === currentRequestIdRef.current) {
                console.error(
                    'Error fetching inference data:',
                    err
                );
                setInferenceError(
                    `Failed to fetch inference data: ${
                        err instanceof Error
                            ? err.message
                            : 'Unknown error'
                    }`
                );
                setInferenceData(null);
            }
        } finally {
            // Only update loading state if this is still the most recent request
            if (requestId === currentRequestIdRef.current) {
                setInferenceLoading(false);
            }
        }
    }, [selectedCompetition, blockHeight]);

    // Function to fetch heights
    const fetchHeights = useCallback(async () => {
        if (!selectedCompetition) return;

        // Generate a unique request ID
        const requestId = ++currentRequestIdRef.current;

        try {
            setHeightsLoading(true);
            const response = await fetchTopicHeights(
                selectedCompetition.topic_id.toString()
            );

            // Check if this is still the most recent request
            if (requestId !== currentRequestIdRef.current) {
                console.log(
                    'Ignoring stale heights response for topic_id:',
                    selectedCompetition.topic_id
                );
                return; // Ignore stale responses
            }

            if (response.data && response.data.heights) {
                setAvailableHeights(response.data.heights);
                // If heights are empty, mark as no data
                if (response.data.heights.length === 0) {
                    setHasNoData(true);
                } else {
                    // If we have heights, make sure we're not in a "no data" state
                    setHasNoData(false);

                    // If we're at the latest height (blockHeight is null), trigger a fetch
                    if (blockHeight === null) {
                        // We'll let the blockHeight effect trigger the fetch
                        // But we need to ensure inferenceData is cleared if we're refreshing
                        if (
                            inferenceData &&
                            inferenceData.data.topic_id !==
                                selectedCompetition.topic_id.toString()
                        ) {
                            setInferenceData(null);
                        }
                    }
                }
            } else {
                // Handle empty heights
                setAvailableHeights([]);
                setHasNoData(true);
            }
        } catch (err) {
            // Only update state if this is still the most recent request
            if (requestId === currentRequestIdRef.current) {
                console.error(
                    'Error fetching heights:',
                    err
                );
                setAvailableHeights([]);
                setHasNoData(true);
            }
        } finally {
            // Only update loading state if this is still the most recent request
            if (requestId === currentRequestIdRef.current) {
                setHeightsLoading(false);
            }
        }
    }, [selectedCompetition, blockHeight, inferenceData]);

    // 데이터 새로고침 함수 - Latest 버튼을 클릭했을 때 사용
    const refreshData = useCallback(() => {
        console.log('Manually refreshing data...');
        // Reset request tracking
        currentRequestIdRef.current = 0;
        // Reset error state
        setInferenceError(null);
        // Reset no data flag
        setHasNoData(false);
        // Fetch fresh data
        fetchInferenceData();
    }, [fetchInferenceData]);

    // Toggle show all heights
    const toggleShowAllHeights = () => {
        setShowAllHeights(!showAllHeights);
    };

    // Handle competition selection
    const handleSelectCompetition = (
        competition: Competition
    ) => {
        // If selecting the same competition, just refresh the data
        if (selectedCompetition?.id === competition.id) {
            console.log(
                'Re-selecting the same competition, refreshing data...'
            );
            // Reset request tracking
            currentRequestIdRef.current = 0;
            // Clear any error state
            setInferenceError(null);
            // Reset no data flag
            setHasNoData(false);
            // Trigger a fresh data fetch
            fetchHeights().then(() => {
                // Force a refresh of inference data after heights are fetched
                if (!hasNoData) {
                    fetchInferenceData();
                }
            });
            return;
        }

        // Reset request tracking when changing competitions
        currentRequestIdRef.current = 0;

        setSelectedCompetition(competition);
        setBlockHeight(null);
        setHasNoData(false); // Reset no data state when changing competition
        setInferenceData(null); // Clear previous inference data
        setInferenceError(null); // Clear previous errors
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
                // Fetch both active and inactive competitions
                const [activeData, inactiveData] =
                    await Promise.all([
                        fetchActiveCompetitions(),
                        fetchInactiveCompetitions(),
                    ]);

                console.log(
                    'Fetched active competitions:',
                    activeData
                );
                console.log(
                    'Fetched inactive competitions:',
                    inactiveData
                );

                setActiveCompetitions(activeData);
                setInactiveCompetitions(inactiveData);

                // Set the first active competition as selected by default
                // If no active competitions, use the first inactive one
                if (activeData.length > 0) {
                    setSelectedCompetition(activeData[0]);
                } else if (inactiveData.length > 0) {
                    setSelectedCompetition(inactiveData[0]);
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

    // Combined effect for fetching data when competition changes
    useEffect(() => {
        // Only fetch data if we have a selected competition
        if (selectedCompetition) {
            // First fetch heights
            fetchHeights();
        }
    }, [selectedCompetition, fetchHeights]);

    // Separate effect for fetching inference data when blockHeight changes
    // Note: We removed selectedCompetition from the dependency array
    useEffect(() => {
        if (selectedCompetition && !hasNoData) {
            // Add a small delay to ensure state updates have propagated
            const timer = setTimeout(() => {
                console.log(
                    'Fetching inference data for topic_id:',
                    selectedCompetition.topic_id,
                    'blockHeight:',
                    blockHeight
                );
                fetchInferenceData();
            }, 100);

            return () => clearTimeout(timer);
        }
    }, [
        blockHeight,
        fetchInferenceData,
        hasNoData,
        selectedCompetition,
    ]);

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
                    activeCompetitions={activeCompetitions}
                    inactiveCompetitions={
                        inactiveCompetitions
                    }
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
                    ) : inferenceError ? (
                        <div className="text-center text-red-500 p-8 bg-red-50 rounded-lg border border-red-200">
                            <div className="text-4xl mb-2">
                                ⚠️
                            </div>
                            <p className="font-medium">
                                {inferenceError}
                            </p>
                            {selectedCompetition &&
                                !selectedCompetition.is_active && (
                                    <p className="mt-2 text-gray-600">
                                        This competition is
                                        inactive and may not
                                        have available data.
                                    </p>
                                )}
                        </div>
                    ) : hasNoData ? (
                        <div className="text-center text-amber-600 p-8 bg-amber-50 rounded-lg border border-amber-200">
                            <div className="text-4xl mb-2">
                                📊
                            </div>
                            <p className="font-medium">
                                No data available for{' '}
                                {selectedCompetition?.name}{' '}
                                (Topic ID:{' '}
                                {
                                    selectedCompetition?.topic_id
                                }
                                )
                            </p>
                            {selectedCompetition &&
                                !selectedCompetition.is_active && (
                                    <p className="mt-2 text-gray-600">
                                        This competition is
                                        inactive and may not
                                        have available data.
                                    </p>
                                )}
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
                                    onRefresh={refreshData}
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
                            {selectedCompetition ? (
                                <p>
                                    No inference data
                                    available for this
                                    competition
                                </p>
                            ) : (
                                <p>
                                    Please select a
                                    competition to view data
                                </p>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
