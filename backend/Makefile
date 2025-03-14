.PHONY: build run test clean

# 기본 변수
BINARY_NAME=allora-monitor
BUILD_DIR=bin

# 빌드 설정
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# 기본 타겟
all: build

# 의존성 설치
deps:
	@echo "Installing dependencies..."
	go mod download

# 빌드
build: deps
	@echo "Building..."
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/app/main.go

# 실행
run:
	@echo "Running..."
	go run cmd/app/main.go

# 테스트
test:
	@echo "Testing..."
	go test ./... -v

# 코드 포맷팅 및 정적 분석
lint:
	@echo "Linting..."
	go fmt ./...
	go vet ./...

# 정리
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)

# 도움말
help:
	@echo "Make targets:"
	@echo "  all    - Default target, builds the application"
	@echo "  deps   - Install dependencies"
	@echo "  build  - Build the application"
	@echo "  run    - Run the application"
	@echo "  test   - Run tests"
	@echo "  lint   - Run code formatting and static analysis"
	@echo "  clean  - Remove build artifacts"
	@echo "  help   - Show this help message" 