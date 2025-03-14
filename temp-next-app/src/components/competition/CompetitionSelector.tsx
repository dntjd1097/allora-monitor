import React from 'react';
import { Button } from '@/components/ui/button';
import { Competition } from '@/lib/api';
import { formatDate } from '@/lib/utils';

interface CompetitionSelectorProps {
    competitions: Competition[];
    selectedCompetition: Competition | null;
    onSelectCompetition: (competition: Competition) => void;
}

export const CompetitionSelector: React.FC<
    CompetitionSelectorProps
> = ({
    competitions,
    selectedCompetition,
    onSelectCompetition,
}) => {
    // Style for better text readability on mobile
    const textStyle = {
        color: '#000000',
        fontWeight: 600,
    };

    return (
        <div className="grid grid-cols-1 md:grid-cols-4 border-b border-gray-300">
            <div className="col-span-1 border-r border-gray-300 p-5 bg-gray-50">
                <h2
                    className="text-xl font-semibold mb-4 text-[var(--foreground)]"
                    style={textStyle}
                >
                    Competition
                </h2>
                <div className="flex flex-wrap gap-2">
                    {competitions.map((competition) => (
                        <Button
                            key={competition.id}
                            variant={
                                selectedCompetition?.id ===
                                competition.id
                                    ? 'default'
                                    : 'outline'
                            }
                            onClick={() =>
                                onSelectCompetition(
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
                    ))}
                </div>
            </div>

            <div className="col-span-3 p-4 bg-white">
                {selectedCompetition && (
                    <div>
                        <h2
                            className="text-xl font-semibold text-indigo-700 mb-2"
                            style={textStyle}
                        >
                            {selectedCompetition.name}
                        </h2>

                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
                            <div className="border border-gray-300 rounded-lg shadow-sm p-4 bg-white hover:shadow-md transition-shadow">
                                <div>
                                    <h3
                                        className="text-center font-medium text-[var(--foreground)] mb-2"
                                        style={textStyle}
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
                                        style={textStyle}
                                    >
                                        Start Date
                                    </h3>
                                    <p
                                        className="text-sm text-center font-medium"
                                        style={textStyle}
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
                                        style={textStyle}
                                    >
                                        End Date
                                    </h3>
                                    <p
                                        className="text-sm text-center font-medium"
                                        style={textStyle}
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
    );
};
