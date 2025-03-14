import React, { useState } from 'react';
import {
    Column,
    SortDirection,
    SortField,
    SynthesisValueItem,
} from '@/types';
import { ColumnSettings } from './ColumnSettings';
import {
    formatLargeNumber,
    formatPercentiles,
    getRankColor,
    getScoreColor,
} from '@/lib/formatters';

interface SynthesisTableProps {
    synthesisValues: SynthesisValueItem[];
}

export const SynthesisTable: React.FC<
    SynthesisTableProps
> = ({ synthesisValues }) => {
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
    const [showColumnMenu, setShowColumnMenu] =
        useState(false);

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
        if (!synthesisValues) {
            return [];
        }

        const values = [...synthesisValues];

        return values.sort((a, b) => {
            let aValue: string | number;
            let bValue: string | number;

            switch (sortField) {
                case 'rank':
                    aValue = a.leaderboard
                        ? parseInt(a.leaderboard.rank)
                        : Number.MAX_SAFE_INTEGER;
                    bValue = b.leaderboard
                        ? parseInt(b.leaderboard.rank)
                        : Number.MAX_SAFE_INTEGER;
                    break;
                case 'username':
                    aValue = a.leaderboard
                        ? a.leaderboard.username.toLowerCase()
                        : 'zzz';
                    bValue = b.leaderboard
                        ? b.leaderboard.username.toLowerCase()
                        : 'zzz';
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
                    aValue = a.leaderboard
                        ? a.leaderboard.loss
                        : Number.MAX_SAFE_INTEGER;
                    bValue = b.leaderboard
                        ? b.leaderboard.loss
                        : Number.MAX_SAFE_INTEGER;
                    break;
                case 'score':
                    aValue = a.leaderboard
                        ? a.leaderboard.score
                        : Number.MIN_SAFE_INTEGER;
                    bValue = b.leaderboard
                        ? b.leaderboard.score
                        : Number.MIN_SAFE_INTEGER;
                    break;
                case 'points':
                    aValue = a.leaderboard
                        ? a.leaderboard.points
                        : 0;
                    bValue = b.leaderboard
                        ? b.leaderboard.points
                        : 0;
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
                return item.leaderboard ? (
                    <span
                        className={getRankColor(
                            item.leaderboard.rank
                        )}
                    >
                        {item.leaderboard.rank}
                    </span>
                ) : (
                    <span>N/A</span>
                );
            case 'username':
                return item.leaderboard ? (
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
                ) : (
                    <div>
                        <div className="flex items-center">
                            <span className="mr-1 text-gray-500">
                                ●
                            </span>
                            <span className="font-medium">
                                Unknown
                            </span>
                        </div>
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
                return item.leaderboard
                    ? item.leaderboard.loss.toFixed(5)
                    : 'N/A';
            case 'score':
                return item.leaderboard ? (
                    <span
                        className={getScoreColor(
                            item.leaderboard.score
                        )}
                    >
                        {item.leaderboard.score.toFixed(5)}
                    </span>
                ) : (
                    <span>N/A</span>
                );
            case 'points':
                return item.leaderboard
                    ? item.leaderboard.points.toFixed(2)
                    : 'N/A';
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

    if (!synthesisValues || synthesisValues.length === 0) {
        return (
            <div className="text-center text-[var(--muted-foreground)] p-8 bg-gray-50 rounded-lg">
                <div className="text-4xl mb-2">📊</div>
                No synthesis values available
            </div>
        );
    }

    return (
        <>
            <h2 className="text-xl font-semibold mb-4 flex items-center text-[var(--foreground)] border-b pb-2 border-gray-200">
                <span className="mr-2">
                    Synthesis Values
                </span>
                <span className="text-sm text-[var(--muted-foreground)] font-normal">
                    ({synthesisValues.length} contributors)
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
                            strokeWidth={2}
                            d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"
                        />
                    </svg>
                    Drag column headers to reorder
                </span>
                <ColumnSettings
                    columns={columns}
                    setColumns={setColumns}
                    showColumnMenu={showColumnMenu}
                    setShowColumnMenu={setShowColumnMenu}
                />
            </div>
            <div className="overflow-x-auto bg-white rounded-lg shadow-sm border border-gray-200 mb-8">
                <table className="w-full border-collapse text-sm">
                    <thead>
                        <tr className="bg-gradient-to-r from-gray-100 to-gray-50">
                            {columns
                                .filter(
                                    (column) =>
                                        column.visible
                                )
                                .map((column) => (
                                    <th
                                        key={column.id}
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
                                                    (c) =>
                                                        c.id ===
                                                        column.id
                                                )
                                            )
                                        }
                                        onDragOver={(e) =>
                                            handleDragOver(
                                                e,
                                                columns.findIndex(
                                                    (c) =>
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
                                ))}
                        </tr>
                    </thead>
                    <tbody>
                        {getSortedSynthesisValues().map(
                            (item, index) => (
                                <tr
                                    key={item.worker}
                                    className={
                                        index % 2 === 0
                                            ? 'bg-gray-50'
                                            : ''
                                    }
                                >
                                    {columns
                                        .filter(
                                            (column) =>
                                                column.visible
                                        )
                                        .map((column) => (
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
                                        ))}
                                </tr>
                            )
                        )}
                    </tbody>
                </table>
            </div>
        </>
    );
};
