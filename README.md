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

## Docker Compose Setup with External Data Volume

This project uses Docker Compose to run the backend, frontend, and nginx services. The backend service uses a named volume for data persistence.

### Prerequisites

-   Docker
-   Docker Compose

### Running the Application

1. Clone the repository:

    ```bash
    git clone <repository-url>
    cd allora-monitor
    ```

2. Start the services:

    ```bash
    docker-compose up -d
    ```

3. Stop the services:
    ```bash
    docker-compose down
    ```

### Data Persistence

The application uses a named volume `allora_monitor_data` for database persistence. This volume is managed by Docker and persists even if the containers are removed.

-   The data is stored in the volume `allora_monitor_data`
-   The volume is mounted to `/app/data` in the backend container

### Backing Up Data

To back up the data, you can use Docker's volume commands:

```bash
# Create a backup
docker run --rm -v allora_monitor_data:/data -v $(pwd):/backup alpine tar -czf /backup/allora_monitor_data_backup.tar.gz /data

# Restore from backup
docker run --rm -v allora_monitor_data:/data -v $(pwd):/backup alpine sh -c "cd /data && tar -xzf /backup/allora_monitor_data_backup.tar.gz --strip 1"
```

### Troubleshooting

If you encounter a "readonly database" error, it might be due to permission issues. Try the following:

1. Check the permissions in the container:

    ```bash
    docker exec -it allora-monitor-backend sh -c "ls -la /app/data"
    ```

2. If needed, fix permissions:

    ```bash
    docker exec -it allora-monitor-backend sh -c "chmod -R 777 /app/data"
    ```

3. Restart the container:
    ```bash
    docker-compose restart backend
    ```

## Development

For local development without Docker:

1. Set up the backend:

    ```bash
    cd backend
    go mod download
    go run cmd/app/main.go
    ```

2. Set up the frontend:
    ```bash
    cd temp-next-app
    npm install
    npm run dev
    ```

Note: When developing locally, the data will be stored in the `backend/data` directory, which is ignored by git.

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

### Production Deployment

For production deployment, we use a separate Docker Compose configuration and deployment scripts that handle external volume creation and management.

#### Using the Deployment Scripts

1. For Linux/macOS:

    ```bash
    # Make the script executable
    chmod +x deploy-production.sh

    # Run the deployment script
    ./deploy-production.sh
    ```

2. For Windows:
    ```cmd
    # Run the deployment script
    deploy-production.bat
    ```

These scripts will:

-   Create the external Docker volume if it doesn't exist
-   Set proper permissions on the volume
-   Pull the latest changes from git (if in a git repository)
-   Build and start the containers using the production configuration
-   Verify that all containers are running

#### Manual Production Deployment

If you prefer to deploy manually:

1. Create the external volume:

    ```bash
    docker volume create allora_monitor_data
    ```

2. Set proper permissions on the volume:

    ```bash
    docker run --rm -v allora_monitor_data:/data alpine sh -c "chmod -R 777 /data"
    ```

3. Deploy using the production Docker Compose file:
    ```bash
    docker-compose -f docker-compose.prod.yml up -d --build
    ```

### Troubleshooting Production Deployment

If you encounter the "readonly database" error in production:

1. Check the volume permissions:

    ```bash
    docker run --rm -v allora_monitor_data:/data alpine sh -c "ls -la /data"
    ```

2. Fix permissions if needed:

    ```bash
    docker run --rm -v allora_monitor_data:/data alpine sh -c "chmod -R 777 /data"
    ```

3. Restart the backend container:
    ```bash
    docker-compose -f docker-compose.prod.yml restart backend
    ```

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
