// Define sorting types
export type SortField =
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
export type SortDirection = 'asc' | 'desc';

// Define column type
export type Column = {
    id: SortField;
    label: string;
    width?: string;
    align?: 'left' | 'center' | 'right';
    visible: boolean;
};

// Define synthesis value item type
export type SynthesisValueItem = {
    worker: string;
    inferer_values: string;
    one_out_inferer_values: string;
    weight: string;
    confidential_percentiles?: string;
    leaderboard?: {
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
export type Pagination = {
    current_height: string;
    next_height: string | null;
    prev_height: string | null;
};
