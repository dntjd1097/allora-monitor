# Allora Monitor

Allora Monitor는 프론트엔드(Next.js)와 백엔드(Go) 서비스로 구성된 모니터링 애플리케이션입니다.

## 프로젝트 구조

```
allora-monitor/
├── .github/
│   └── workflows/       # GitHub Actions 워크플로우
├── backend/             # Go 백엔드 서비스
├── temp-next-app/       # Next.js 프론트엔드 서비스
├── caddy/               # Caddy 웹 서버 설정
│   └── Caddyfile        # Caddy 설정 파일
└── docker-compose.yml   # Docker Compose 설정 파일
```

## 로컬 개발 환경 설정

### 필수 요구 사항

-   Docker
-   Docker Compose

### 로컬에서 실행하기

1. 저장소 클론:

```bash
git clone https://github.com/your-username/allora-monitor.git
cd allora-monitor
```

2. Docker Compose로 서비스 실행:

```bash
docker compose up -d
```

3. 서비스 접근:
    - 웹 인터페이스: http://localhost
    - API: http://localhost/api

## 배포

### GitHub Actions를 사용한 자동 배포

이 프로젝트는 GitHub Actions를 사용하여 자동으로 배포됩니다. `main` 브랜치에 변경 사항이 푸시되면 자동으로 배포 워크플로우가 실행됩니다.

### 수동 배포

1. 서버에 SSH로 접속
2. 저장소 클론 또는 풀:

```bash
git clone https://github.com/your-username/allora-monitor.git
# 또는
cd allora-monitor && git pull
```

3. Docker Compose로 서비스 실행:

```bash
docker compose up -d --build
```

## 서비스 구성

### 백엔드 (Go)

-   포트: 8080 (내부)
-   API 엔드포인트: `/api/*`

### 프론트엔드 (Next.js)

-   포트: 3000 (내부)
-   웹 인터페이스: `/`

### Caddy 웹 서버

-   포트: 80, 443
-   자동 HTTPS 지원
-   리버스 프록시 설정:
    -   `/` -> 프론트엔드
    -   `/api/*` -> 백엔드
