#!/bin/bash

# 색상 설정
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}  Docker Swarm 문제 해결 스크립트   ${NC}"
echo -e "${BLUE}=====================================${NC}"

# Docker 실행 중인지 확인
echo -e "\n${YELLOW}[1/7] Docker 상태 확인 중...${NC}"
if ! docker info &>/dev/null; then
    echo -e "${RED}[오류] Docker가 실행되고 있지 않습니다.${NC}"
    echo "Docker 서비스를 시작하세요:"
    echo "sudo systemctl start docker (Linux)"
    echo "Docker Desktop 애플리케이션 시작 (macOS/Windows)"
    exit 1
else
    echo -e "${GREEN}[성공] Docker가 실행 중입니다.${NC}"
fi

# Docker Swarm 모드 활성화 확인
echo -e "\n${YELLOW}[2/7] Docker Swarm 모드 확인 중...${NC}"
if ! docker info | grep -q "Swarm: active"; then
    echo -e "${RED}[오류] Docker Swarm 모드가 활성화되어 있지 않습니다.${NC}"
    echo "Swarm 모드를 초기화하세요:"
    echo "docker swarm init"
    exit 1
else
    echo -e "${GREEN}[성공] Docker Swarm 모드가 활성화되어 있습니다.${NC}"
fi

# 네트워크 확인
echo -e "\n${YELLOW}[3/7] Docker 네트워크 확인 중...${NC}"
if ! docker network ls | grep -q "allora-web-network"; then
    echo -e "${RED}[오류] allora-web-network 네트워크가 존재하지 않습니다.${NC}"
    echo "네트워크를 생성하세요:"
    echo "docker network create --driver overlay --attachable allora-web-network"
else
    NETWORK_DRIVER=$(docker network inspect allora-web-network -f '{{.Driver}}')
    if [ "$NETWORK_DRIVER" != "overlay" ]; then
        echo -e "${RED}[오류] allora-web-network 네트워크가 overlay 타입이 아닙니다. 현재 타입: $NETWORK_DRIVER${NC}"
        echo "기존 네트워크를 제거하고 overlay 타입으로 다시 생성하세요:"
        echo "docker network rm allora-web-network"
        echo "docker network create --driver overlay --attachable allora-web-network"
    else
        echo -e "${GREEN}[성공] allora-web-network 네트워크가 overlay 타입으로 설정되어 있습니다.${NC}"
    fi
fi

# 이미지 존재 확인
echo -e "\n${YELLOW}[4/7] Docker 이미지 확인 중...${NC}"
if ! docker images | grep -q "allora-monitor-backend"; then
    echo -e "${RED}[오류] allora-monitor-backend 이미지가 존재하지 않습니다.${NC}"
    echo "이미지를 빌드하세요:"
    echo "docker compose build backend"
else
    echo -e "${GREEN}[성공] allora-monitor-backend 이미지가 존재합니다.${NC}"
fi

if ! docker images | grep -q "allora-monitor-frontend"; then
    echo -e "${RED}[오류] allora-monitor-frontend 이미지가 존재하지 않습니다.${NC}"
    echo "이미지를 빌드하세요:"
    echo "docker compose build frontend"
else
    echo -e "${GREEN}[성공] allora-monitor-frontend 이미지가 존재합니다.${NC}"
fi

# 볼륨 확인
echo -e "\n${YELLOW}[5/7] Docker 볼륨 확인 중...${NC}"
if ! docker volume ls | grep -q "allora_monitor_data"; then
    echo -e "${RED}[오류] allora_monitor_data 볼륨이 존재하지 않습니다.${NC}"
    echo "볼륨을 생성하세요:"
    echo "docker volume create allora_monitor_data"
else
    echo -e "${GREEN}[성공] allora_monitor_data 볼륨이 존재합니다.${NC}"
fi

# 서비스 확인
echo -e "\n${YELLOW}[6/7] Docker 서비스 확인 중...${NC}"
if docker service ls | grep -q "allora-monitor"; then
    echo -e "${GREEN}[정보] 다음 서비스가 실행 중입니다:${NC}"
    docker service ls | grep "allora-monitor"
    
    # 서비스 로그 확인
    echo -e "\n${YELLOW}[7/7] 서비스 로그 확인${NC}"
    echo -e "${BLUE}[프론트엔드 서비스 로그]${NC}"
    docker service logs allora-monitor_frontend --tail 10 2>/dev/null || echo -e "${RED}[오류] 프론트엔드 서비스 로그를 가져올 수 없습니다.${NC}"
    
    echo -e "\n${BLUE}[백엔드 서비스 로그]${NC}"
    docker service logs allora-monitor_backend --tail 10 2>/dev/null || echo -e "${RED}[오류] 백엔드 서비스 로그를 가져올 수 없습니다.${NC}"
else
    echo -e "${YELLOW}[정보] 실행 중인 allora-monitor 서비스가 없습니다.${NC}"
fi

# 배포 명령어 제안
echo -e "\n${BLUE}===== 배포 방법 =====${NC}"
echo -e "1. 이미지 빌드:\n   docker compose build"
echo -e "2. 스크립트로 배포:\n   ./deploy-swarm.sh"
echo -e "3. 수동으로 배포:\n   docker stack deploy -c docker-compose.swarm.yml allora-monitor"
echo -e "4. 서비스 상태 확인:\n   docker service ls"
echo -e "5. 서비스 스케일 조정:\n   docker service scale allora-monitor_frontend=5"

echo -e "\n${BLUE}===== 포트 접근 정보 =====${NC}"
echo -e "프론트엔드: http://localhost:5000"
echo -e "백엔드 API: http://localhost:5001"

echo -e "\n${BLUE}===== 문제 해결 =====${NC}"
echo -e "1. 스택 제거:\n   docker stack rm allora-monitor"
echo -e "2. 네트워크 재설정:\n   docker network rm allora-web-network && docker network create --driver overlay --attachable allora-web-network"
echo -e "3. 로그 확인:\n   docker service logs allora-monitor_frontend" 