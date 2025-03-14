FROM node:20-alpine AS builder

WORKDIR /app

# 의존성 설치
COPY package.json bun.lockb ./
RUN npm install

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN npm run build

# 실행 이미지
FROM node:20-alpine

WORKDIR /app

# 빌드된 파일과 필요한 파일 복사
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./

# 환경 변수 설정
ENV NODE_ENV=production

# 포트 노출
EXPOSE 3000

# 실행
CMD ["npm", "start"] 