FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# 의존성 복사 및 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 애플리케이션 빌드
RUN go build -o allora-monitor ./cmd/app/main.go

# 실행 이미지
FROM alpine:latest

WORKDIR /app

# 빌드된 바이너리 복사
COPY --from=builder /app/allora-monitor .
COPY --from=builder /app/config.json .
COPY --from=builder /app/configs ./configs

# 필요한 디렉토리 생성
RUN mkdir -p /app/data

# 포트 노출
EXPOSE 8080

# 실행
CMD ["./allora-monitor"] 