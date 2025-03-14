import React from 'react';
import { Column } from '@/types';

interface ColumnSettingsProps {
    columns: Column[];
    setColumns: React.Dispatch<
        React.SetStateAction<Column[]>
    >;
    showColumnMenu: boolean;
    setShowColumnMenu: React.Dispatch<
        React.SetStateAction<boolean>
    >;
}

export const ColumnSettings: React.FC<
    ColumnSettingsProps
> = ({
    columns,
    setColumns,
    showColumnMenu,
    setShowColumnMenu,
}) => {
    return (
        <div className="relative">
            <button
                className="px-3 py-1 bg-indigo-50 text-indigo-600 rounded-md text-sm font-medium hover:bg-indigo-100 flex items-center"
                onClick={(e) => {
                    e.stopPropagation();
                    setShowColumnMenu(!showColumnMenu);
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
                        strokeWidth={2}
                        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
                    />
                    <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                    />
                </svg>
                Column settings
            </button>

            {showColumnMenu && (
                <div className="absolute right-0 mt-2 w-64 bg-white rounded-md shadow-lg z-10 border border-gray-200">
                    <div className="p-3">
                        <h3 className="text-sm font-medium text-[var(--foreground)] mb-2">
                            Select columns to display
                        </h3>
                        <div className="space-y-2">
                            {columns.map((column) => (
                                <div
                                    key={column.id}
                                    className="flex items-center"
                                >
                                    <input
                                        type="checkbox"
                                        id={`column-${column.id}`}
                                        checked={
                                            column.visible
                                        }
                                        onChange={(e) => {
                                            setColumns(
                                                columns.map(
                                                    (col) =>
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
                                        {column.label}
                                    </label>
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};
