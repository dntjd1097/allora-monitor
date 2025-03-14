import React from 'react';
import { Pagination } from '@/types';

interface PaginationControlsProps {
    pagination: Pagination | null;
    blockHeight: string | null;
    setBlockHeight: (height: string | null) => void;
    availableHeights: string[];
    heightsLoading: boolean;
    customHeight: string;
    setCustomHeight: (height: string) => void;
    showAllHeights: boolean;
    toggleShowAllHeights: () => void;
    onRefresh?: () => void;
}

export const PaginationControls: React.FC<
    PaginationControlsProps
> = ({
    pagination,
    blockHeight,
    setBlockHeight,
    availableHeights,
    heightsLoading,
    customHeight,
    setCustomHeight,
    showAllHeights,
    toggleShowAllHeights,
    onRefresh,
}) => {
    // Go to previous block
    const goToPrevBlock = () => {
        if (pagination?.prev_height) {
            setBlockHeight(pagination.prev_height);
        }
    };

    // Go to next block
    const goToNextBlock = () => {
        if (pagination?.next_height) {
            setBlockHeight(pagination.next_height);
        }
    };

    // Go to latest block (set block height to null)
    const goToLatestBlock = () => {
        if (blockHeight !== null || customHeight !== '') {
            setBlockHeight(null);
            setCustomHeight('');
        } else if (onRefresh) {
            onRefresh();
        }
    };

    // Handle height selection change
    const handleHeightChange = (
        e: React.ChangeEvent<HTMLSelectElement>
    ) => {
        const selectedHeight = e.target.value;
        setCustomHeight('');
        if (selectedHeight !== 'latest') {
            setBlockHeight(selectedHeight);
        }
        goToLatestBlock();
    };

    if (!pagination) return null;

    // Latest 버튼의 활성화 상태 결정
    const isLatestActive =
        blockHeight === null && customHeight === '';

    return (
        <div className="mb-8">
            <div className="flex flex-col md:flex-row items-start md:items-center justify-between bg-gray-50 p-4 rounded-lg border border-gray-200">
                <div className="flex items-center space-x-2 mb-3 md:mb-0">
                    <div className="flex flex-col space-y-2">
                        <div className="flex items-center space-x-2">
                            <label
                                htmlFor="height-select"
                                className="text-sm font-medium text-[var(--foreground)]"
                            >
                                Select Height:
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
                                disabled={heightsLoading}
                            >
                                <option value="latest">
                                    Latest
                                </option>

                                {(showAllHeights
                                    ? availableHeights
                                    : availableHeights.slice(
                                          0,
                                          10
                                      )
                                ).map((height) => (
                                    <option
                                        key={height}
                                        value={height}
                                    >
                                        {height}
                                    </option>
                                ))}
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
                    </div>
                </div>
                <div className="flex flex-col md:flex-row md:items-center space-y-3 md:space-y-0 md:space-x-4 w-full md:w-auto">
                    <div className="flex items-center space-x-2">
                        <button
                            onClick={goToPrevBlock}
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
                                    strokeWidth={2}
                                    d="M15 19l-7-7 7-7"
                                />
                            </svg>
                            Previous
                        </button>
                        <button
                            onClick={goToLatestBlock}
                            className={`px-3 py-1 rounded text-sm font-medium flex items-center ${
                                isLatestActive
                                    ? 'bg-indigo-200 text-indigo-700 hover:bg-indigo-300'
                                    : 'bg-indigo-50 text-indigo-600 hover:bg-indigo-100'
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
                                    strokeWidth={2}
                                    d="M13 10V3L4 14h7v7l9-11h-7z"
                                />
                            </svg>
                            Latest
                        </button>
                        <button
                            onClick={goToNextBlock}
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
                                    strokeWidth={2}
                                    d="M9 5l7 7-7 7"
                                />
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};
