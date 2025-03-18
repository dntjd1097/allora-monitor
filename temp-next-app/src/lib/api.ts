import axios, { AxiosError } from 'axios';

// Backend API base URL - get from environment variable or use default
const API_BASE_URL =
    process.env.NEXT_PUBLIC_API_URL ||
    'https://backend.allora-inference.kro.kr';

// Create axios instance with default config
const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    // Add timeout to prevent hanging requests
    timeout: 1000,
});

// Maximum number of retries
const MAX_RETRIES = 0;

// Helper function to handle API requests with retry logic
async function apiRequest<T>(
    requestFn: () => Promise<T>,
    retries = MAX_RETRIES
): Promise<T> {
    try {
        return await requestFn();
    } catch (error) {
        if (retries > 0) {
            console.log(
                `Request failed, retrying... (${
                    MAX_RETRIES - retries + 1
                }/${MAX_RETRIES})`
            );
            // Wait before retrying (exponential backoff)
            await new Promise((resolve) =>
                setTimeout(
                    resolve,
                    1000 * (MAX_RETRIES - retries + 1)
                )
            );
            return apiRequest(requestFn, retries - 1);
        }

        // Format error message for better debugging
        let errorMessage = 'Unknown error occurred';
        if (error instanceof AxiosError) {
            if (error.code === 'ECONNABORTED') {
                errorMessage =
                    'Request timeout - server took too long to respond';
            } else if (error.code === 'ERR_NETWORK') {
                errorMessage =
                    'Network error - cannot connect to the backend server. Make sure the server is running at ' +
                    API_BASE_URL;
            } else if (error.response) {
                errorMessage = `Server error: ${error.response.status} ${error.response.statusText}`;
            }
        }

        console.error('API request failed:', errorMessage);
        throw error;
    }
}

// API endpoints
export const endpoints = {
    health: '/api/health',
    competitions: '/api/competitions',
    competitionsRange: '/api/competitions/range',
    competitionsV2: '/api/competitions/v2',
    stats: '/api/stats',
    activeTopics: '/api/topics/active',
    topicInference: '/api/topics/inference',
    allTopicInferences: '/api/topics/inferences',
    topicStats: '/api/topics/stats',
    topicHeights: '/api/topics/heights',
};

// Interface for competition data
export interface Competition {
    id: number;
    name: string;
    preview_image_url: string | null;
    description: string | null;
    detailed_description?: string | null;
    topic_id: number;
    prize_pool: number;
    start_date: string;
    end_date: string;
    season_id: number;
    tags: string[];
    is_active: boolean;
}

// Interface for inference data
export interface TopicInferenceData {
    data: {
        confidence_interval_raw_percentiles: string[];
        confidence_interval_values: string[];
        inference_block_height: string;
        loss_block_height: string;
        network_inferences: {
            combined_value: string;
            naive_value: string;
            reputer: string;
            reputer_request_nonce: {
                reputer_nonce: {
                    block_height: string;
                };
            };
            synthesis_value: Array<{
                inferer_values: string;
                leaderboard: {
                    cosmos_address: string;
                    first_name: string;
                    last_name: string;
                    is_active: boolean;
                    loss: number;
                    points: number;
                    rank: string;
                    score: number;
                    username: string;
                };
                one_out_inferer_values: string;
                weight: string;
                worker: string;
                confidential_percentiles?: string;
            }>;
            extra_data: Record<string, unknown> | null;
            forecaster_values: string[];
            one_in_forecaster_values: string[];
            one_out_forecaster_values: string[];
            one_out_inferer_forecaster_values: string[];
        };
        timestamp: string;
        topic_id: string;
    };
    pagination: {
        current_height: string;
        next_height: string | null;
        prev_height: string | null;
    };
    status: string;
}

export interface SynthesisValue {
    inferer_values: string;
    leaderboard: {
        cosmos_address: string;
        first_name: string;
        is_active: boolean;
        last_name: string;
        loss: number;
        points: number;
        rank: string;
        score: number;
        username: string;
    };
    one_out_inferer_values: string;
    weight: string;
    worker: string;
    confidential_percentiles?: string;
}

// Mock data for development when backend is not available
const MOCK_DATA = {
    health: {
        status: 'healthy',
        monitor_status: 'running',
        timestamp: new Date().toISOString(),
    },
    competitions: [
        {
            id: 1,
            name: '12h ETH/USD Volatility Prediction',
            preview_image_url:
                'https://allora-forge-assets.s3.us-east-1.amazonaws.com/forge-5.jpeg',
            description: null,
            detailed_description:
                '# 12h ETH/USD Volatility Prediction\n\nWelcome to our 12-hour Ethereum Volatility Modeling Challenge!',
            topic_id: 28,
            prize_pool: 5000,
            start_date: '2025-01-16T23:00:00Z',
            end_date: '2025-02-20T18:00:00Z',
            season_id: 1,
            tags: [],
            is_active: true,
        },
        {
            id: 2,
            name: 'BTC Price Prediction',
            preview_image_url:
                'https://allora-forge-assets.s3.us-east-1.amazonaws.com/forge-6.jpeg',
            description: null,
            detailed_description:
                '# BTC Price Prediction Challenge',
            topic_id: 29,
            prize_pool: 3000,
            start_date: '2025-01-10T23:00:00Z',
            end_date: '2025-02-10T18:00:00Z',
            season_id: 1,
            tags: [],
            is_active: true,
        },
    ],
    competitionsV2: [
        {
            id: 6,
            name: '5 min BTC/USD Price Prediction',
            preview_image_url:
                'https://allora-forge-assets.s3.us-east-1.amazonaws.com/forge-3.jpeg',
            description: null,
            topic_id: 47,
            prize_pool: 0,
            start_date: '2025-02-28T15:00:00Z',
            end_date: '2025-03-14T18:00:00Z',
            season_id: 1,
            tags: [],
            is_active: true,
        },
        {
            id: 7,
            name: '6h BTC/USD Volatility Prediction (updating every 5 min)',
            preview_image_url:
                'https://allora-forge-assets.s3.us-east-1.amazonaws.com/forge-1.png',
            description: null,
            topic_id: 50,
            prize_pool: 0,
            start_date: '2025-03-03T15:00:00Z',
            end_date: '2025-03-17T18:00:00Z',
            season_id: 1,
            tags: [],
            is_active: true,
        },
    ],
    stats: {
        total_records: 0,
        oldest_record: new Date().toISOString(),
        newest_record: new Date().toISOString(),
        database_size_bytes: 0,
        average_record_size_bytes: 0,
        monitor_running: true,
        last_fetch_time: new Date().toISOString(),
        fetch_interval_minutes: 60,
    },
    activeTopics: [],
    topicInferences: [],
    topicStats: [],
    topicHeights: [],
};

// Environment variable to enable mock data (can be set in .env)
const USE_MOCK_DATA =
    process.env.NEXT_PUBLIC_USE_MOCK_DATA === 'true';

// Debug mode
const DEBUG = process.env.NEXT_PUBLIC_DEBUG === 'true';

// API functions
export const fetchHealth = async () => {
    if (DEBUG) console.log('Fetching health data');
    if (USE_MOCK_DATA) return MOCK_DATA.health;

    return apiRequest(async () => {
        const response = await api.get(endpoints.health);
        return response.data;
    });
};

export const fetchCompetitions = async () => {
    if (DEBUG) console.log('Fetching competitions data');
    if (USE_MOCK_DATA) return MOCK_DATA.competitions;

    return apiRequest(async () => {
        const response = await api.get(
            endpoints.competitions
        );
        return response.data;
    });
};

export const fetchCompetitionsByTimeRange = async (
    start?: string,
    end?: string
) => {
    if (DEBUG)
        console.log('Fetching competitions by time range', {
            start,
            end,
        });
    if (USE_MOCK_DATA) return MOCK_DATA.competitions;

    return apiRequest(async () => {
        const params = new URLSearchParams();
        if (start) params.append('start', start);
        if (end) params.append('end', end);

        const url = `${
            endpoints.competitionsRange
        }?${params.toString()}`;
        const response = await api.get(url);
        return response.data;
    });
};

export const fetchCompetitionsV2 = async (
    params?: Record<string, string>
) => {
    if (DEBUG)
        console.log('Fetching competitions v2', params);
    if (USE_MOCK_DATA) return MOCK_DATA.competitionsV2;

    return apiRequest(async () => {
        const response = await api.get(
            endpoints.competitionsV2,
            { params }
        );
        return response.data;
    });
};

// Fetch active competitions from competitions/v2 endpoint
export const fetchActiveCompetitions = async (): Promise<
    Competition[]
> => {
    if (DEBUG)
        console.log(
            'Fetching active competitions from competitions/v2'
        );
    if (USE_MOCK_DATA) return MOCK_DATA.competitionsV2;

    return apiRequest(async () => {
        const response = await api.get(
            `${endpoints.competitionsV2}?active=true`
        );
        // The API returns an array directly, not wrapped in active_competitions
        return response.data || [];
    });
};

// Fetch inactive competitions from competitions/v2 endpoint
export const fetchInactiveCompetitions = async (): Promise<
    Competition[]
> => {
    if (DEBUG)
        console.log(
            'Fetching inactive competitions from competitions/v2'
        );
    if (USE_MOCK_DATA) return [];

    return apiRequest(async () => {
        const response = await api.get(
            `${endpoints.competitionsV2}?active=false`
        );
        // The API returns an array directly
        return response.data || [];
    });
};

// New function to fetch active topics from competitions/v2 endpoint
export const fetchActiveTopicsV2 = async () => {
    if (DEBUG)
        console.log(
            'Fetching active topics from competitions/v2'
        );
    if (USE_MOCK_DATA) return MOCK_DATA.activeTopics;

    return apiRequest(async () => {
        const response = await api.get(
            `${endpoints.competitionsV2}/topics`
        );
        return response.data || [];
    });
};

export const fetchStats = async () => {
    if (DEBUG) console.log('Fetching stats data');
    if (USE_MOCK_DATA) return MOCK_DATA.stats;

    return apiRequest(async () => {
        const response = await api.get(endpoints.stats);
        return response.data;
    });
};

export const fetchActiveTopics = async () => {
    if (DEBUG) console.log('Fetching active topics');
    if (USE_MOCK_DATA) return MOCK_DATA.activeTopics;

    return apiRequest(async () => {
        const response = await api.get(
            endpoints.activeTopics
        );
        return response.data;
    });
};

export const fetchTopicInference = async (
    topicId: string,
    blockHeight?: string
): Promise<TopicInferenceData | null> => {
    if (DEBUG)
        console.log('Fetching topic inference', {
            topicId,
            blockHeight,
        });
    if (USE_MOCK_DATA) {
        // Return mock data for topic inference
        return {
            data: {
                confidence_interval_raw_percentiles: [
                    '2.28',
                    '15.87',
                    '50',
                    '84.13',
                    '97.72',
                ],
                confidence_interval_values: [
                    '81023.93392288291362209543709286858',
                    '81052.50224402173715210230095786263',
                    '81103.51429626812341052786668863302',
                    '81156.78211001620875157122560046372',
                    '81211.76310530350545492787280433983',
                ],
                inference_block_height:
                    blockHeight || '2986625',
                loss_block_height: '2986520',
                network_inferences: {
                    combined_value:
                        '81107.09719329704363325372738870811',
                    naive_value:
                        '81107.09719329708041912988702951235',
                    reputer:
                        'allo1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqas6usy',
                    reputer_request_nonce: {
                        reputer_nonce: {
                            block_height: '2986641',
                        },
                    },
                    synthesis_value: [
                        {
                            inferer_values: '85099.74',
                            leaderboard: {
                                cosmos_address:
                                    'allo109cu4dlhj7cr46jkfuvhag8eu9xzccscj0x5ed',
                                first_name: 'Chau Nguyen',
                                is_active: true,
                                last_name: 'Dinh',
                                loss: 7.20993,
                                points: 4558.563,
                                rank: '53',
                                score: 0.000021011645211021,
                                username: 'ch_und',
                            },
                            one_out_inferer_values:
                                '81106.98927401564433172701230829321',
                            weight: '0.00002256908063366186229843693104187439',
                            worker: 'allo109cu4dlhj7cr46jkfuvhag8eu9xzccscj0x5ed',
                            confidential_percentiles:
                                '84.13~97.72',
                        },
                        {
                            inferer_values:
                                '81091.33346178834',
                            leaderboard: {
                                cosmos_address:
                                    'allo1qqzwc5d7hnhj5jgx5jvc9ltk78dz4wu0vj0hnl',
                                first_name: 'Ava',
                                is_active: true,
                                last_name: 'Johnson',
                                loss: 2.45678,
                                points: 3560.123,
                                rank: '12',
                                score: 0.0456789,
                                username: 'avajohnson',
                            },
                            one_out_inferer_values:
                                '81134.80324300089401170200432694453',
                            weight: '0.1234567890123456789012345678901234',
                            worker: 'allo1qqzwc5d7hnhj5jgx5jvc9ltk78dz4wu0vj0hnl',
                            confidential_percentiles:
                                '50~84.13',
                        },
                        {
                            inferer_values:
                                '81091.33346178834',
                            leaderboard: {
                                cosmos_address:
                                    'allo1asdfjkl5d7hnhj5jgx5jvc9ltk78dz4wu0vjasd',
                                first_name: 'Michael',
                                is_active: false,
                                last_name: 'Smith',
                                loss: 5.6789,
                                points: 2890.456,
                                rank: '25',
                                score: 0.0234567,
                                username: 'mikesmith',
                            },
                            one_out_inferer_values:
                                '81134.80324300089401170200432694453',
                            weight: '0.0987654321098765432109876543210987',
                            worker: 'allo1asdfjkl5d7hnhj5jgx5jvc9ltk78dz4wu0vjasd',
                            confidential_percentiles:
                                '2.28~15.87',
                        },
                        {
                            inferer_values:
                                '81091.33346178834',
                            leaderboard: {
                                cosmos_address:
                                    'allo1zxcvbnm5d7hnhj5jgx5jvc9ltk78dz4wu0vzxcv',
                                first_name: 'Sarah',
                                is_active: true,
                                last_name: 'Williams',
                                loss: 1.23456,
                                points: 4120.789,
                                rank: '8',
                                score: 0.0678901,
                                username: 'sarahw',
                            },
                            one_out_inferer_values:
                                '81134.80324300089401170200432694453',
                            weight: '0.3456789012345678901234567890123456',
                            worker: 'allo1zxcvbnm5d7hnhj5jgx5jvc9ltk78dz4wu0vzxcv',
                            confidential_percentiles:
                                '97.72~99.87',
                        },
                        {
                            inferer_values:
                                '81218.46052581798',
                            leaderboard: {
                                cosmos_address:
                                    'allo1e2l4c3uj4xcvcznhfa2rnxwzd5h4sgqahg6g50',
                                first_name: 'Soham',
                                is_active: true,
                                last_name: 'Mukherjee',
                                loss: 4.79198,
                                points: 7656.128,
                                rank: '20',
                                score: 0.0691081177953054,
                                username: 'Dingon',
                            },
                            one_out_inferer_values:
                                '81099.19543633386302861319100230013',
                            weight: '0.1077306416254524851330940853219689',
                            worker: 'allo1e2l4c3uj4xcvcznhfa2rnxwzd5h4sgqahg6g50',
                        },
                        {
                            inferer_values:
                                '81156.84065964638',
                            leaderboard: {
                                cosmos_address:
                                    'allo1daf27k7ftl3k9huhrc0unr2t52ll3eprd02h2v',
                                first_name: 'Van',
                                is_active: false,
                                last_name: 'Tran',
                                loss: 3.6,
                                points: 8992.082,
                                rank: '1',
                                score: -0.0323554750737416,
                                username: 'Ciddy',
                            },
                            one_out_inferer_values:
                                '81105.71014255035939166802763575500',
                            weight: '0.1080771498255803768334328836651266',
                            worker: 'allo1daf27k7ftl3k9huhrc0unr2t52ll3eprd02h2v',
                        },
                        {
                            inferer_values:
                                '81113.7793287524',
                            leaderboard: {
                                cosmos_address:
                                    'allo1569mc4yk4vn4uzkf4c2yldx5nz6l02rkw0hjs5',
                                first_name: 'Jeremy',
                                is_active: true,
                                last_name: 'Huynh',
                                loss: 4.03521,
                                points: 7704.273,
                                rank: '19',
                                score: 0.0120801684563554,
                                username: 'MekongLabs',
                            },
                            one_out_inferer_values:
                                '81106.37578702838762766217466903088',
                            weight: '0.1563373016676767994916178472878381',
                            worker: 'allo1569mc4yk4vn4uzkf4c2yldx5nz6l02rkw0hjs5',
                        },
                        {
                            inferer_values:
                                '81091.33346178834',
                            leaderboard: {
                                cosmos_address:
                                    'allo1qwerty5d7hnhj5jgx5jvc9ltk78dz4wu0vqwer',
                                first_name: 'John',
                                is_active: true,
                                last_name: 'Doe',
                                loss: 3.45678,
                                points: 3250.789,
                                rank: '18',
                                score: 0.0345678,
                                username: 'johndoe',
                            },
                            one_out_inferer_values:
                                '81134.80324300089401170200432694453',
                            weight: '0.1765432109876543210987654321098765',
                            worker: 'allo1qwerty5d7hnhj5jgx5jvc9ltk78dz4wu0vqwer',
                            confidential_percentiles:
                                '0.13~2.28',
                        },
                    ],
                    extra_data: null,
                    forecaster_values: [],
                    one_in_forecaster_values: [],
                    one_out_forecaster_values: [],
                    one_out_inferer_forecaster_values: [],
                },
                timestamp: '2025-03-14T08:48:55+09:00',
                topic_id: topicId,
            },
            pagination: {
                current_height: '2986625',
                next_height: null,
                prev_height: '2986590',
            },
            status: 'success',
        };
    }

    try {
        return await apiRequest(async () => {
            const params = new URLSearchParams();
            params.append('topic_id', topicId);
            if (blockHeight) {
                params.append('height', blockHeight);
            }

            const url = `${
                endpoints.topicInference
            }?${params.toString()}`;
            const response = await api.get(url);
            return response.data;
        });
    } catch (error) {
        // Check if it's a 404 error (data not found)
        if (
            error instanceof AxiosError &&
            error.response?.status === 404
        ) {
            console.warn(
                `No inference data available for topic ${topicId}`
            );
            return null;
        }
        // Re-throw other errors
        throw error;
    }
};

export const fetchAllTopicInferences = async () => {
    if (DEBUG) console.log('Fetching all topic inferences');
    if (USE_MOCK_DATA) return MOCK_DATA.topicInferences;

    return apiRequest(async () => {
        const response = await api.get(
            endpoints.allTopicInferences
        );
        return response.data;
    });
};

export const fetchTopicStats = async () => {
    if (DEBUG) console.log('Fetching topic stats');
    if (USE_MOCK_DATA) return MOCK_DATA.topicStats;

    return apiRequest(async () => {
        const response = await api.get(
            endpoints.topicStats
        );
        return response.data;
    });
};

export const fetchTopicHeights = async (
    topicId: string
) => {
    if (DEBUG)
        console.log(
            'Fetching topic heights for topic_id:',
            topicId
        );
    if (USE_MOCK_DATA) return MOCK_DATA.topicHeights;

    try {
        return await apiRequest(async () => {
            const params = new URLSearchParams();
            params.append('topic_id', topicId);

            const url = `${
                endpoints.topicHeights
            }?${params.toString()}`;
            const response = await api.get(url);

            // Check if heights is null and convert to empty array
            if (
                response.data &&
                response.data.data &&
                response.data.data.heights === null
            ) {
                response.data.data.heights = [];
            }

            return response.data;
        });
    } catch (error) {
        console.error(
            `Error fetching heights for topic ${topicId}:`,
            error
        );
        // Return a valid response with empty heights
        return {
            data: {
                heights: [],
                heights_count: 0,
                limit: 100,
                offset: 0,
                topic_id: topicId,
                total_count: 0,
            },
            status: 'success',
        };
    }
};
