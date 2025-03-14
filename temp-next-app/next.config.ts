import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
    output: 'standalone',
    typescript: {
        // !! WARN !!
        // 프로덕션 환경에서는 타입 오류를 무시하지 않는 것이 좋습니다.
        // 여기서는 빌드를 위해 임시로 설정합니다.
        ignoreBuildErrors: true,
    },
    eslint: {
        // 빌드 시 ESLint 검사를 건너뜁니다.
        ignoreDuringBuilds: true,
    },
    // HTML에 환경 변수 스크립트 추가
    async headers() {
        return [
            {
                source: '/:path*',
                headers: [
                    {
                        key: 'Cache-Control',
                        value: 'no-store, must-revalidate',
                    },
                ],
            },
        ];
    },
};

export default nextConfig;
