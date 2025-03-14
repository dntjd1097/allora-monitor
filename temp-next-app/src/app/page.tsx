'use client';

import { useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { Loading } from '@/components/ui/loading';
import {
    Competition,
    fetchActiveCompetitions,
    fetchTopicInference,
    fetchTopicHeights,
    TopicInferenceData,
} from '@/lib/api';
import { formatDate } from '@/lib/utils';

// Define sorting types
type SortField =
    | 'rank'
    | 'username'
    | 'address'
    | 'inferer_values'
    | 'weight'
    | 'one_out_values'
    | 'loss'
    | 'score'
    | 'points'
    | 'confidential_percentiles';
type SortDirection = 'asc' | 'desc';

// Define column type
type Column = {
    id: SortField;
    label: string;
    width?: string;
    align?: 'left' | 'center' | 'right';
    visible: boolean;
};

// Define synthesis value item type
type SynthesisValueItem = {
    worker: string;
    inferer_values: string;
    one_out_inferer_values: string;
    weight: string;
    confidential_percentiles?: string;
    leaderboard: {
        rank: string;
        username: string;
        first_name?: string;
        last_name?: string;
        is_active: boolean;
        loss: number;
        score: number;
        points: number;
    };
};

// Define pagination type
type Pagination = {
    current_height: string;
    next_height: string | null;
    prev_height: string | null;
};

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

    // Sorting state
    const [sortField, setSortField] =
        useState<SortField>('rank');
    const [sortDirection, setSortDirection] =
        useState<SortDirection>('asc');

    // Column order state
    const [columns, setColumns] = useState<Column[]>([
        {
            id: 'rank',
            label: 'Rank',
            align: 'center',
            visible: true,
        },
        {
            id: 'username',
            label: 'Worker Name',
            align: 'left',
            visible: true,
        },
        {
            id: 'address',
            label: 'Address',
            align: 'left',
            width: 'max-w-[150px]',
            visible: true,
        },
        {
            id: 'inferer_values',
            label: 'Inferer Values',
            align: 'right',
            visible: true,
        },
        {
            id: 'weight',
            label: 'Weight',
            align: 'right',
            visible: true,
        },
        {
            id: 'one_out_values',
            label: 'One Out Values',
            align: 'right',
            visible: true,
        },
        {
            id: 'loss',
            label: 'Loss',
            align: 'right',
            visible: true,
        },
        {
            id: 'score',
            label: 'Score',
            align: 'right',
            visible: true,
        },
        {
            id: 'points',
            label: 'Points',
            align: 'right',
            visible: true,
        },
        {
            id: 'confidential_percentiles',
            label: 'Confidential Percentiles',
            align: 'center',
            visible: true,
        },
    ]);

    // Drag and drop state
    const [draggedColumnIndex, setDraggedColumnIndex] =
        useState<number | null>(null);

    // Handle column drag start
    const handleDragStart = (index: number) => {
        setDraggedColumnIndex(index);
    };

    // Handle column drag over
    const handleDragOver = (
        e: React.DragEvent,
        index: number
    ) => {
        e.preventDefault();
        if (draggedColumnIndex === null) return;

        // Don't do anything if dragging over the same column
        if (draggedColumnIndex === index) return;

        // Reorder columns
        const newColumns = [...columns];
        const draggedColumn =
            newColumns[draggedColumnIndex];
        newColumns.splice(draggedColumnIndex, 1);
        newColumns.splice(index, 0, draggedColumn);

        setColumns(newColumns);
        setDraggedColumnIndex(index);
    };

    // Handle column drag end
    const handleDragEnd = () => {
        setDraggedColumnIndex(null);
    };

    // 페이지네이션 상태
    const [blockHeight, setBlockHeight] = useState<
        string | null
    >(null);
    const [pagination, setPagination] =
        useState<Pagination | null>(null);
    const [availableHeights, setAvailableHeights] =
        useState<string[]>([]);
    const [heightsLoading, setHeightsLoading] =
        useState<boolean>(false);
    const [customHeight, setCustomHeight] =
        useState<string>('');
    const [showAllHeights, setShowAllHeights] =
        useState<boolean>(false);

    // 모바일 환경에서 텍스트 가독성 향상을 위한 스타일
    const textStyle = {
        color: '#000000',
        fontWeight: 600,
    };

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

    useEffect(() => {
        const fetchInferenceData = async () => {
            if (!selectedCompetition) return;

            try {
                setInferenceLoading(true);
                // Use the blockHeight parameter directly with the API function
                const data = await fetchTopicInference(
                    selectedCompetition.topic_id.toString(),
                    blockHeight || undefined
                );
                console.log(
                    'Fetched inference data:',
                    data
                );
                setInferenceData(data);

                // 페이지네이션 정보 설정
                if (data.pagination) {
                    setPagination(data.pagination);
                    // 현재 블록 높이 업데이트
                    setBlockHeight(
                        data.pagination.current_height
                    );
                }
            } catch (err) {
                console.error(
                    'Error fetching inference data:',
                    err
                );
            } finally {
                setInferenceLoading(false);
            }
        };

        fetchInferenceData();
    }, [selectedCompetition, blockHeight]);

    useEffect(() => {
        const fetchHeights = async () => {
            if (!selectedCompetition) return;

            try {
                setHeightsLoading(true);
                const response = await fetchTopicHeights(
                    selectedCompetition.topic_id.toString()
                );
                if (
                    response.data &&
                    response.data.heights
                ) {
                    setAvailableHeights(
                        response.data.heights
                    );
                }
            } catch (err) {
                console.error(
                    'Error fetching heights:',
                    err
                );
            } finally {
                setHeightsLoading(false);
            }
        };

        fetchHeights();
    }, [selectedCompetition]);

    const handleSelectCompetition = (
        competition: Competition
    ) => {
        setSelectedCompetition(competition);
        setBlockHeight(null);
    };

    // Helper function to format large numbers
    const formatLargeNumber = (value: string) => {
        if (!value) return 'N/A';
        // Truncate to 8 digits for display
        return value.substring(0, 8);
    };

    // Helper function to format percentiles for better display
    const formatPercentiles = (
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
    const getRankColor = (rank: string) => {
        const rankNum = parseInt(rank);
        if (rankNum <= 10)
            return 'text-green-600 font-bold';
        if (rankNum <= 30) return 'text-blue-600';
        if (rankNum <= 50) return 'text-orange-600';
        return 'text-[var(--foreground)]';
    };

    // Helper function to get color based on score
    const getScoreColor = (score: number) => {
        if (score > 0.05) return 'text-green-600';
        if (score > 0) return 'text-blue-600';
        if (score > -0.05) return 'text-orange-600';
        return 'text-red-600';
    };

    // Handle sorting
    const handleSort = (field: SortField) => {
        if (sortField === field) {
            // Toggle direction if same field
            setSortDirection(
                sortDirection === 'asc' ? 'desc' : 'asc'
            );
        } else {
            // Set new field and default to ascending
            setSortField(field);
            setSortDirection('asc');
        }
    };

    // Get sorted synthesis values
    const getSortedSynthesisValues = () => {
        if (
            !inferenceData?.data?.network_inferences
                ?.synthesis_value
        ) {
            return [];
        }

        const values = [
            ...inferenceData.data.network_inferences
                .synthesis_value,
        ];

        return values.sort((a, b) => {
            let aValue: string | number;
            let bValue: string | number;

            switch (sortField) {
                case 'rank':
                    aValue = parseInt(a.leaderboard.rank);
                    bValue = parseInt(b.leaderboard.rank);
                    break;
                case 'username':
                    aValue =
                        a.leaderboard.username.toLowerCase();
                    bValue =
                        b.leaderboard.username.toLowerCase();
                    break;
                case 'address':
                    aValue = a.worker.toLowerCase();
                    bValue = b.worker.toLowerCase();
                    break;
                case 'inferer_values':
                    aValue = parseFloat(a.inferer_values);
                    bValue = parseFloat(b.inferer_values);
                    break;
                case 'weight':
                    aValue = parseFloat(a.weight);
                    bValue = parseFloat(b.weight);
                    break;
                case 'one_out_values':
                    aValue = parseFloat(
                        a.one_out_inferer_values
                    );
                    bValue = parseFloat(
                        b.one_out_inferer_values
                    );
                    break;
                case 'loss':
                    aValue = a.leaderboard.loss;
                    bValue = b.leaderboard.loss;
                    break;
                case 'score':
                    aValue = a.leaderboard.score;
                    bValue = b.leaderboard.score;
                    break;
                case 'points':
                    aValue = a.leaderboard.points;
                    bValue = b.leaderboard.points;
                    break;
                case 'confidential_percentiles':
                    aValue =
                        (a.confidential_percentiles as string) ||
                        '';
                    bValue =
                        (b.confidential_percentiles as string) ||
                        '';
                    break;
                default:
                    return 0;
            }

            // Handle NaN values
            if (typeof aValue === 'number' && isNaN(aValue))
                aValue = 0;
            if (typeof bValue === 'number' && isNaN(bValue))
                bValue = 0;

            // Sort direction
            const direction =
                sortDirection === 'asc' ? 1 : -1;

            // Compare
            if (aValue < bValue) return -1 * direction;
            if (aValue > bValue) return 1 * direction;
            return 0;
        });
    };

    // Render sort indicator
    const renderSortIndicator = (field: SortField) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? ' ▲' : ' ▼';
    };

    // Get cell content based on column id
    const getCellContent = (
        item: SynthesisValueItem,
        columnId: SortField
    ) => {
        switch (columnId) {
            case 'rank':
                return (
                    <span
                        className={getRankColor(
                            item.leaderboard.rank
                        )}
                    >
                        {item.leaderboard.rank}
                    </span>
                );
            case 'username':
                return (
                    <div>
                        <div className="flex items-center">
                            {item.leaderboard.is_active ? (
                                <span className="mr-1 text-green-500">
                                    ●
                                </span>
                            ) : (
                                <span className="mr-1 text-red-500">
                                    ●
                                </span>
                            )}
                            <span className="font-medium">
                                {item.leaderboard.username}
                            </span>
                        </div>
                        {item.leaderboard.first_name && (
                            <div className="text-xs text-[var(--muted-foreground)]">
                                {
                                    item.leaderboard
                                        .first_name
                                }{' '}
                                {item.leaderboard.last_name}
                            </div>
                        )}
                    </div>
                );
            case 'address':
                return (
                    <div
                        className="tooltip"
                        title={item.worker}
                    >
                        {item.worker.substring(0, 12)}...
                    </div>
                );
            case 'inferer_values':
                return formatLargeNumber(
                    item.inferer_values
                );
            case 'weight':
                return parseFloat(item.weight).toFixed(8);
            case 'one_out_values':
                return formatLargeNumber(
                    item.one_out_inferer_values
                );
            case 'loss':
                return item.leaderboard.loss.toFixed(5);
            case 'score':
                return (
                    <span
                        className={getScoreColor(
                            item.leaderboard.score
                        )}
                    >
                        {item.leaderboard.score.toFixed(5)}
                    </span>
                );
            case 'points':
                return item.leaderboard.points.toFixed(2);
            case 'confidential_percentiles':
                return formatPercentiles(
                    item.confidential_percentiles
                );
            default:
                return null;
        }
    };

    // Get cell class based on column id
    const getCellClass = (columnId: SortField) => {
        const column = columns.find(
            (col) => col.id === columnId
        );
        let classes = 'border border-gray-300 p-2';

        if (column?.align === 'center')
            classes += ' text-center';
        if (column?.align === 'right')
            classes += ' text-right';
        if (column?.id === 'address')
            classes += ' font-mono text-xs truncate';
        if (
            column?.id === 'inferer_values' ||
            column?.id === 'one_out_values' ||
            column?.id === 'weight'
        )
            classes += ' font-mono';

        return classes;
    };

    // 컬럼 설정 메뉴 표시 상태
    const [showColumnMenu, setShowColumnMenu] =
        useState(false);

    // 이전 블록으로 이동
    const goToPrevBlock = () => {
        if (pagination?.prev_height) {
            setBlockHeight(pagination.prev_height);
        }
    };

    // 다음 블록으로 이동
    const goToNextBlock = () => {
        if (pagination?.next_height) {
            setBlockHeight(pagination.next_height);
        }
    };

    // 최신 블록으로 이동 (블록 높이 null로 설정)
    const goToLatestBlock = () => {
        setBlockHeight(null);
    };

    // Block Height 선택 핸들러
    const handleHeightChange = (
        e: React.ChangeEvent<HTMLSelectElement>
    ) => {
        const selectedHeight = e.target.value;
        if (selectedHeight === 'custom') {
            // 사용자 정의 입력 필드로 전환
            return;
        }
        setBlockHeight(
            selectedHeight === 'latest'
                ? null
                : selectedHeight
        );
        setCustomHeight('');
    };

    // 사용자 정의 Height 입력 핸들러
    const handleCustomHeightChange = (
        e: React.ChangeEvent<HTMLInputElement>
    ) => {
        setCustomHeight(e.target.value);
    };

    // 사용자 정의 Height 제출 핸들러
    const handleCustomHeightSubmit = (
        e: React.FormEvent
    ) => {
        e.preventDefault();
        if (customHeight.trim()) {
            setBlockHeight(customHeight.trim());
        }
    };

    // 모든 heights 표시 토글
    const toggleShowAllHeights = () => {
        setShowAllHeights(!showAllHeights);
    };

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
                <div className="grid grid-cols-1 md:grid-cols-4 border-b border-gray-300">
                    <div className="col-span-1 border-r border-gray-300 p-5 bg-gray-50">
                        <h2
                            className="text-xl font-semibold mb-4 text-[var(--foreground)]"
                            style={textStyle}
                        >
                            Competition
                        </h2>
                        <div className="flex flex-wrap gap-2">
                            {competitions.map(
                                (competition) => (
                                    <Button
                                        key={competition.id}
                                        variant={
                                            selectedCompetition?.id ===
                                            competition.id
                                                ? 'default'
                                                : 'outline'
                                        }
                                        onClick={() =>
                                            handleSelectCompetition(
                                                competition
                                            )
                                        }
                                        className={
                                            selectedCompetition?.id ===
                                            competition.id
                                                ? 'bg-indigo-600 hover:bg-indigo-700 text-white'
                                                : 'hover:bg-gray-100'
                                        }
                                    >
                                        {competition.id}
                                    </Button>
                                )
                            )}
                        </div>
                    </div>

                    <div className="col-span-3 p-4 bg-white">
                        {selectedCompetition && (
                            <div>
                                <h2
                                    className="text-xl font-semibold text-indigo-700 mb-2"
                                    style={textStyle}
                                >
                                    {
                                        selectedCompetition.name
                                    }
                                </h2>

                                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
                                    <div className="border border-gray-300 rounded-lg shadow-sm p-4 bg-white hover:shadow-md transition-shadow">
                                        <div>
                                            <h3
                                                className="text-center font-medium text-[var(--foreground)] mb-2"
                                                style={
                                                    textStyle
                                                }
                                            >
                                                Topic ID
                                            </h3>
                                            <p className="text-2xl font-bold text-center text-purple-600">
                                                {
                                                    selectedCompetition.topic_id
                                                }
                                            </p>
                                        </div>
                                    </div>

                                    <div className="border border-gray-300 rounded-lg shadow-sm p-4 bg-white hover:shadow-md transition-shadow">
                                        <div>
                                            <h3
                                                className="text-center font-medium text-[var(--foreground)] mb-2"
                                                style={
                                                    textStyle
                                                }
                                            >
                                                Start Date
                                            </h3>
                                            <p
                                                className="text-sm text-center font-medium"
                                                style={
                                                    textStyle
                                                }
                                            >
                                                {formatDate(
                                                    selectedCompetition.start_date,
                                                    'PPpp'
                                                )}
                                            </p>
                                        </div>
                                    </div>

                                    <div className="border border-gray-300 rounded-lg shadow-sm p-4 bg-white hover:shadow-md transition-shadow">
                                        <div>
                                            <h3
                                                className="text-center font-medium text-[var(--foreground)] mb-2"
                                                style={
                                                    textStyle
                                                }
                                            >
                                                End Date
                                            </h3>
                                            <p
                                                className="text-sm text-center font-medium"
                                                style={
                                                    textStyle
                                                }
                                            >
                                                {formatDate(
                                                    selectedCompetition.end_date,
                                                    'PPpp'
                                                )}
                                            </p>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>

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
                            <h2
                                className="text-xl font-semibold mb-4 text-[var(--foreground)] border-b pb-2 border-gray-200"
                                style={textStyle}
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
                                                inferenceData
                                                    .data
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
                                                inferenceData
                                                    .data
                                                    .network_inferences
                                                    .naive_value
                                            )}
                                        </p>
                                    </div>
                                </div>

                                <div className="border border-gray-300 rounded-lg shadow-sm hover:shadow-md transition-shadow p-4 bg-white">
                                    <div>
                                        <h3 className="text-center font-medium text-[var(--foreground)] mb-2">
                                            Inference Block
                                            Height
                                        </h3>
                                        <p className="text-2xl font-bold text-center text-purple-600">
                                            {
                                                inferenceData
                                                    .data
                                                    .inference_block_height
                                            }
                                        </p>
                                    </div>
                                </div>

                                <div className="border border-gray-300 rounded-lg shadow-sm hover:shadow-md transition-shadow p-4 bg-white">
                                    <div>
                                        <h3 className="text-center font-medium text-[var(--foreground)] mb-2">
                                            Loss Block
                                            Height
                                        </h3>
                                        <p className="text-2xl font-bold text-center text-orange-600">
                                            {
                                                inferenceData
                                                    .data
                                                    .loss_block_height
                                            }
                                        </p>
                                    </div>
                                </div>
                            </div>

                            {/* 페이지네이션 컨트롤 추가 */}
                            {pagination && (
                                <div className="mb-8">
                                    <div className="flex flex-col md:flex-row items-start md:items-center justify-between bg-gray-50 p-4 rounded-lg border border-gray-200">
                                        <div className="flex items-center space-x-2 mb-3 md:mb-0">
                                            <span className="text-sm font-medium text-[var(--foreground)]">
                                                Current
                                                Block
                                                Height:
                                            </span>
                                            <span className="text-sm font-mono bg-white px-2 py-1 rounded border border-gray-300">
                                                {
                                                    pagination.current_height
                                                }
                                            </span>
                                        </div>
                                        <div className="flex flex-col md:flex-row md:items-center space-y-3 md:space-y-0 md:space-x-4 w-full md:w-auto">
                                            <div className="flex flex-col space-y-2">
                                                <div className="flex items-center space-x-2">
                                                    <label
                                                        htmlFor="height-select"
                                                        className="text-sm font-medium text-[var(--foreground)]"
                                                    >
                                                        Select
                                                        Height:
                                                    </label>
                                                    <select
                                                        id="height-select"
                                                        value={
                                                            customHeight
                                                                ? 'custom'
                                                                : blockHeight ||
                                                                  'latest'
                                                        }
                                                        onChange={
                                                            handleHeightChange
                                                        }
                                                        className="text-sm rounded border border-gray-300 px-2 py-1 bg-white"
                                                        disabled={
                                                            heightsLoading
                                                        }
                                                    >
                                                        <option value="latest">
                                                            Latest
                                                        </option>
                                                        <option value="custom">
                                                            Custom...
                                                        </option>
                                                        {(showAllHeights
                                                            ? availableHeights
                                                            : availableHeights.slice(
                                                                  0,
                                                                  10
                                                              )
                                                        ).map(
                                                            (
                                                                height
                                                            ) => (
                                                                <option
                                                                    key={
                                                                        height
                                                                    }
                                                                    value={
                                                                        height
                                                                    }
                                                                >
                                                                    {
                                                                        height
                                                                    }
                                                                </option>
                                                            )
                                                        )}
                                                    </select>
                                                    {availableHeights.length >
                                                        10 && (
                                                        <button
                                                            type="button"
                                                            onClick={
                                                                toggleShowAllHeights
                                                            }
                                                            className="text-xs text-indigo-600 hover:text-indigo-800"
                                                        >
                                                            {showAllHeights
                                                                ? 'Show Less'
                                                                : `Show All (${availableHeights.length})`}
                                                        </button>
                                                    )}
                                                </div>
                                                {heightsLoading && (
                                                    <span className="text-xs text-[var(--muted-foreground)]">
                                                        Loading
                                                        heights...
                                                    </span>
                                                )}
                                                <form
                                                    onSubmit={
                                                        handleCustomHeightSubmit
                                                    }
                                                    className="flex items-center space-x-2"
                                                >
                                                    <input
                                                        type="text"
                                                        value={
                                                            customHeight
                                                        }
                                                        onChange={
                                                            handleCustomHeightChange
                                                        }
                                                        placeholder="Enter block height"
                                                        className="text-sm rounded border border-gray-300 px-2 py-1 bg-white"
                                                    />
                                                    <button
                                                        type="submit"
                                                        className="px-2 py-1 bg-indigo-50 text-indigo-600 rounded text-sm font-medium hover:bg-indigo-100"
                                                    >
                                                        Go
                                                    </button>
                                                </form>
                                            </div>
                                            <div className="flex items-center space-x-2">
                                                <button
                                                    onClick={
                                                        goToPrevBlock
                                                    }
                                                    disabled={
                                                        !pagination.prev_height
                                                    }
                                                    className={`px-3 py-1 rounded text-sm font-medium flex items-center ${
                                                        pagination.prev_height
                                                            ? 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100'
                                                            : 'bg-gray-100 text-[var(--muted-foreground)] cursor-not-allowed'
                                                    }`}
                                                >
                                                    <svg
                                                        className="w-4 h-4 mr-1"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        viewBox="0 0 24 24"
                                                        xmlns="http://www.w3.org/2000/svg"
                                                    >
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            strokeWidth={
                                                                2
                                                            }
                                                            d="M15 19l-7-7 7-7"
                                                        />
                                                    </svg>
                                                    Previous
                                                </button>
                                                <button
                                                    onClick={
                                                        goToLatestBlock
                                                    }
                                                    className="px-3 py-1 bg-indigo-50 text-indigo-600 rounded text-sm font-medium hover:bg-indigo-100 flex items-center"
                                                >
                                                    <svg
                                                        className="w-4 h-4 mr-1"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        viewBox="0 0 24 24"
                                                        xmlns="http://www.w3.org/2000/svg"
                                                    >
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            strokeWidth={
                                                                2
                                                            }
                                                            d="M13 10V3L4 14h7v7l9-11h-7z"
                                                        />
                                                    </svg>
                                                    Latest
                                                </button>
                                                <button
                                                    onClick={
                                                        goToNextBlock
                                                    }
                                                    disabled={
                                                        !pagination.next_height
                                                    }
                                                    className={`px-3 py-1 rounded text-sm font-medium flex items-center ${
                                                        pagination.next_height
                                                            ? 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100'
                                                            : 'bg-gray-100 text-[var(--muted-foreground)] cursor-not-allowed'
                                                    }`}
                                                >
                                                    Next
                                                    <svg
                                                        className="w-4 h-4 ml-1"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        viewBox="0 0 24 24"
                                                        xmlns="http://www.w3.org/2000/svg"
                                                    >
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            strokeWidth={
                                                                2
                                                            }
                                                            d="M9 5l7 7-7 7"
                                                        />
                                                    </svg>
                                                </button>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            )}

                            {/* Synthesis Values section first */}
                            {inferenceData.data
                                .network_inferences
                                .synthesis_value &&
                                inferenceData.data
                                    .network_inferences
                                    .synthesis_value
                                    .length > 0 && (
                                    <>
                                        <h2 className="text-xl font-semibold mb-4 flex items-center text-[var(--foreground)] border-b pb-2 border-gray-200">
                                            <span className="mr-2">
                                                Synthesis
                                                Values
                                            </span>
                                            <span className="text-sm text-[var(--muted-foreground)] font-normal">
                                                (
                                                {
                                                    inferenceData
                                                        .data
                                                        .network_inferences
                                                        .synthesis_value
                                                        .length
                                                }{' '}
                                                contributors)
                                            </span>
                                        </h2>
                                        <div className="mb-2 flex justify-between items-center">
                                            <span className="inline-flex items-center text-sm text-[var(--muted-foreground)]">
                                                <svg
                                                    className="w-4 h-4 mr-1"
                                                    fill="none"
                                                    stroke="currentColor"
                                                    viewBox="0 0 24 24"
                                                    xmlns="http://www.w3.org/2000/svg"
                                                >
                                                    <path
                                                        strokeLinecap="round"
                                                        strokeLinejoin="round"
                                                        strokeWidth={
                                                            2
                                                        }
                                                        d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
                                                    />
                                                </svg>
                                                Drag column
                                                headers to
                                                reorder
                                            </span>
                                            <div className="relative">
                                                <button
                                                    className="px-3 py-1 bg-indigo-50 text-indigo-600 rounded-md text-sm font-medium hover:bg-indigo-100 flex items-center"
                                                    onClick={(
                                                        e
                                                    ) => {
                                                        e.stopPropagation();
                                                        setShowColumnMenu(
                                                            !showColumnMenu
                                                        );
                                                    }}
                                                >
                                                    <svg
                                                        className="w-4 h-4 mr-1"
                                                        fill="none"
                                                        stroke="currentColor"
                                                        viewBox="0 0 24 24"
                                                        xmlns="http://www.w3.org/2000/svg"
                                                    >
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            strokeWidth={
                                                                2
                                                            }
                                                            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                                                        />
                                                        <path
                                                            strokeLinecap="round"
                                                            strokeLinejoin="round"
                                                            strokeWidth={
                                                                2
                                                            }
                                                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                                                        />
                                                    </svg>
                                                    Column
                                                    settings
                                                </button>

                                                {showColumnMenu && (
                                                    <div className="absolute right-0 mt-2 w-64 bg-white rounded-md shadow-lg z-10 border border-gray-200">
                                                        <div className="p-3">
                                                            <h3 className="text-sm font-medium text-[var(--foreground)] mb-2">
                                                                Select
                                                                columns
                                                                to
                                                                display
                                                            </h3>
                                                            <div className="space-y-2">
                                                                {columns.map(
                                                                    (
                                                                        column
                                                                    ) => (
                                                                        <div
                                                                            key={
                                                                                column.id
                                                                            }
                                                                            className="flex items-center"
                                                                        >
                                                                            <input
                                                                                type="checkbox"
                                                                                id={`column-${column.id}`}
                                                                                checked={
                                                                                    column.visible
                                                                                }
                                                                                onChange={(
                                                                                    e
                                                                                ) => {
                                                                                    setColumns(
                                                                                        columns.map(
                                                                                            (
                                                                                                col
                                                                                            ) =>
                                                                                                col.id ===
                                                                                                column.id
                                                                                                    ? {
                                                                                                          ...col,
                                                                                                          visible:
                                                                                                              e
                                                                                                                  .target
                                                                                                                  .checked,
                                                                                                      }
                                                                                                    : col
                                                                                        )
                                                                                    );
                                                                                }}
                                                                                className="h-4 w-4 text-indigo-600 rounded border-gray-300"
                                                                            />
                                                                            <label
                                                                                htmlFor={`column-${column.id}`}
                                                                                className="ml-2 text-sm text-[var(--foreground)]"
                                                                            >
                                                                                {
                                                                                    column.label
                                                                                }
                                                                            </label>
                                                                        </div>
                                                                    )
                                                                )}
                                                            </div>
                                                        </div>
                                                    </div>
                                                )}
                                            </div>
                                        </div>
                                        <div className="overflow-x-auto bg-white rounded-lg shadow-sm border border-gray-200 mb-8">
                                            <table className="w-full border-collapse text-sm">
                                                <thead>
                                                    <tr className="bg-gradient-to-r from-gray-100 to-gray-50">
                                                        {columns
                                                            .filter(
                                                                (
                                                                    column
                                                                ) =>
                                                                    column.visible
                                                            )
                                                            .map(
                                                                (
                                                                    column
                                                                ) => (
                                                                    <th
                                                                        key={
                                                                            column.id
                                                                        }
                                                                        className="border border-gray-300 p-2 text-left cursor-move hover:bg-gray-200 font-semibold text-[var(--foreground)]"
                                                                        onClick={() =>
                                                                            handleSort(
                                                                                column.id
                                                                            )
                                                                        }
                                                                        draggable
                                                                        onDragStart={() =>
                                                                            handleDragStart(
                                                                                columns.findIndex(
                                                                                    (
                                                                                        c
                                                                                    ) =>
                                                                                        c.id ===
                                                                                        column.id
                                                                                )
                                                                            )
                                                                        }
                                                                        onDragOver={(
                                                                            e
                                                                        ) =>
                                                                            handleDragOver(
                                                                                e,
                                                                                columns.findIndex(
                                                                                    (
                                                                                        c
                                                                                    ) =>
                                                                                        c.id ===
                                                                                        column.id
                                                                                )
                                                                            )
                                                                        }
                                                                        onDragEnd={
                                                                            handleDragEnd
                                                                        }
                                                                        style={{
                                                                            width: column.width,
                                                                        }}
                                                                    >
                                                                        <div className="flex items-center justify-between">
                                                                            <div className="flex items-center">
                                                                                <span className="mr-1">
                                                                                    <svg
                                                                                        className="w-4 h-4 text-[var(--muted-foreground)]"
                                                                                        fill="none"
                                                                                        stroke="currentColor"
                                                                                        viewBox="0 0 24 24"
                                                                                        xmlns="http://www.w3.org/2000/svg"
                                                                                    >
                                                                                        <path
                                                                                            strokeLinecap="round"
                                                                                            strokeLinejoin="round"
                                                                                            strokeWidth={
                                                                                                2
                                                                                            }
                                                                                            d="M8 9l4-4 4 4m0 6l-4 4-4-4"
                                                                                        />
                                                                                    </svg>
                                                                                </span>
                                                                                {
                                                                                    column.label
                                                                                }
                                                                                {renderSortIndicator(
                                                                                    column.id
                                                                                )}
                                                                            </div>
                                                                        </div>
                                                                    </th>
                                                                )
                                                            )}
                                                    </tr>
                                                </thead>
                                                <tbody>
                                                    {getSortedSynthesisValues().map(
                                                        (
                                                            item,
                                                            index
                                                        ) => (
                                                            <tr
                                                                key={
                                                                    item.worker
                                                                }
                                                                className={
                                                                    index %
                                                                        2 ===
                                                                    0
                                                                        ? 'bg-gray-50'
                                                                        : ''
                                                                }
                                                            >
                                                                {columns
                                                                    .filter(
                                                                        (
                                                                            column
                                                                        ) =>
                                                                            column.visible
                                                                    )
                                                                    .map(
                                                                        (
                                                                            column
                                                                        ) => (
                                                                            <td
                                                                                key={`${item.worker}-${column.id}`}
                                                                                className={getCellClass(
                                                                                    column.id
                                                                                )}
                                                                            >
                                                                                {getCellContent(
                                                                                    item,
                                                                                    column.id
                                                                                )}
                                                                            </td>
                                                                        )
                                                                    )}
                                                            </tr>
                                                        )
                                                    )}
                                                </tbody>
                                            </table>
                                        </div>
                                    </>
                                )}

                            {/* Confidence Intervals and Additional Information grid */}
                            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
                                <div>
                                    <h2 className="text-xl font-semibold mb-4 flex items-center text-[var(--foreground)] border-b pb-2 border-gray-200">
                                        <span className="mr-2">
                                            Confidence
                                            Intervals
                                        </span>
                                        <span className="text-sm text-[var(--muted-foreground)] font-normal">
                                            (Percentiles and
                                            their values)
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
                                                {inferenceData.data.confidence_interval_raw_percentiles.map(
                                                    (
                                                        percentile,
                                                        index
                                                    ) => (
                                                        <tr
                                                            key={
                                                                percentile
                                                            }
                                                            className={
                                                                index %
                                                                    2 ===
                                                                0
                                                                    ? 'bg-gray-50'
                                                                    : ''
                                                            }
                                                        >
                                                            <td className="border border-gray-300 p-2">
                                                                <span className="font-medium">
                                                                    {
                                                                        percentile
                                                                    }

                                                                    %
                                                                </span>
                                                            </td>
                                                            <td className="border border-gray-300 p-2 font-mono">
                                                                {formatLargeNumber(
                                                                    inferenceData
                                                                        .data
                                                                        .confidence_interval_values[
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
                                            Additional
                                            Information
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
                                                        inferenceData
                                                            .data
                                                            .timestamp,
                                                        'PPpp'
                                                    )}
                                                </p>
                                                <p className="text-xs text-[var(--muted-foreground)] mt-1">
                                                    (UTC:{' '}
                                                    {new Date(
                                                        inferenceData.data.timestamp
                                                    ).toISOString()}
                                                    )
                                                </p>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
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
