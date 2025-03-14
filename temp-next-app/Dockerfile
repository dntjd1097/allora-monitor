FROM imbios/bun-node:18-slim AS builder
ARG DEBIAN_FRONTEND=noninteractive

WORKDIR /app

# 종속성 파일 복사
COPY package.json  ./
RUN bun install 

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드 (ESLint 검사 건너뛰기)
RUN NEXT_DISABLE_ESLINT=1 npm run build

# 실행 이미지
FROM node:20-slim AS runner

WORKDIR /app

ENV NODE_ENV=production
# ENV NEXT_TELEMETRY_DISABLED=1
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

# 환경 변수 설정

# 포트 노출
EXPOSE 3000

# 실행
CMD ["node", "server.js"] 