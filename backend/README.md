# Allora Monitor

Allora Monitor는 Allora Network의 경쟁 데이터를 모니터링하고 저장하는 서비스입니다.

## 주요 기능

-   Allora Network API에서 경쟁 데이터를 주기적으로 수집
-   SQLite + JSON + 압축 방식으로 효율적인 데이터 저장
-   수집된 데이터에 접근할 수 있는 REST API 제공
-   데이터베이스 통계 및 모니터링 상태 확인 기능

## 프로젝트 구조

```
.
├── api/            # API 정의 파일
├── cmd/            # 애플리케이션 진입점
│   └── app/        # 메인 애플리케이션
├── configs/        # 설정 파일
├── data/           # 데이터베이스 파일 저장 디렉토리
├── docs/           # 문서
├── internal/       # 내부 패키지
│   └── app/        # 애플리케이션 코어 로직
├── pkg/            # 외부에서 사용 가능한 패키지
│   └── utils/      # 유틸리티 함수
└── test/           # 테스트 파일
```

## 시작하기

### 필수 조건

-   Go 1.16 이상

### 설치

```bash
# 저장소 클론
git clone https://github.com/sonamu/allora-monitor.git
cd allora-monitor

# 의존성 설치
go mod download
```

### 환경 변수 설정

다음 환경 변수를 설정하여 애플리케이션 동작을 제어할 수 있습니다:

-   `SERVER_PORT`: 서버 포트 (기본값: 8080)
-   `LOG_LEVEL`: 로그 레벨 (기본값: info)
-   `DATABASE_PATH`: 데이터베이스 파일 경로 (기본값: data/allora-monitor.db)
-   `ALLORA_API_BASE_URL`: Allora API 기본 URL (기본값: https://forge.allora.network)
-   `API_REQUEST_TIMEOUT_SECONDS`: API 요청 타임아웃 (기본값: 30)
-   `MONITORING_INTERVAL_MINUTES`: 모니터링 간격 (기본값: 60)
-   `DATA_RETENTION_DAYS`: 데이터 보존 기간 (기본값: 90)

### 실행

```bash
# 애플리케이션 실행
go run cmd/app/main.go
```

## API 엔드포인트

-   `GET /`: 기본 정보
-   `GET /health`: 서비스 상태 확인
-   `GET /api/competitions`: 최신 경쟁 데이터 조회
-   `GET /api/stats`: 데이터베이스 통계 및 모니터링 상태 조회

## 빌드

```bash
# 실행 파일 빌드
go build -o bin/allora-monitor cmd/app/main.go
```

## 테스트

```bash
# 모든 테스트 실행
go test ./...
```

## 데이터 저장 방식

Allora Monitor는 SQLite + JSON + 압축 방식을 사용하여 데이터를 효율적으로 저장합니다:

1. 수집된 JSON 데이터를 그대로 유지하여 유연성 확보
2. Snappy 압축 알고리즘을 사용하여 저장 공간 최적화
3. SQLite의 트랜잭션 및 인덱싱 기능을 활용하여 빠른 조회 지원
4. 설정된 보존 기간이 지난 데이터는 자동으로 정리

## 라이센스

이 프로젝트는 MIT 라이센스 하에 배포됩니다.
