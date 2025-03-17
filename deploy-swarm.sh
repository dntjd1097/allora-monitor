#!/bin/bash

# 오류 발생 시 스크립트 중단
set -e

# 스크립트가 실행되는 위치를 저장
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================"
echo "   Allora Monitor Swarm 배포 스크립트   "
echo "========================================"

# 도커 스웜 모드가 초기화되어 있는지 확인
if ! docker info | grep -q "Swarm: active"; then
    echo "도커 스웜 모드 초기화 중..."
    docker swarm init --advertise-addr $(hostname -i) || true
else
    echo "도커 스웜 모드가 이미 활성화되어 있습니다."
fi

# 기존 스택이 있다면 제거 
if docker stack ls | grep -q "allora-monitor"; then
    echo "기존 스택 제거 중..."
    docker stack rm allora-monitor
    
    # 스택이 완전히 제거될 때까지 대기
    echo "스택 제거 대기 중..."
    while docker stack ls | grep -q "allora-monitor"; do
        sleep 2
    done
    
    # 추가적인 정리 시간 부여
    sleep 5
fi

# 네트워크 생성 또는 확인
if ! docker network ls | grep -q "allora-web-network"; then
    echo "오버레이 네트워크 생성 중..."
    docker network create --driver overlay --attachable allora-web-network
else
    # 기존 네트워크가 오버레이 타입인지 확인
    NETWORK_DRIVER=$(docker network inspect allora-web-network -f '{{.Driver}}')
    if [ "$NETWORK_DRIVER" != "overlay" ]; then
        echo "기존 네트워크를 제거하고 오버레이 네트워크로 다시 생성합니다..."
        docker network rm allora-web-network || true
        docker network create --driver overlay --attachable allora-web-network
    else
        echo "오버레이 네트워크가 이미 존재합니다."
    fi
fi

# 데이터 볼륨이 존재하는지 확인하고 없으면 생성
if ! docker volume ls | grep -q "allora_monitor_data"; then
    echo "데이터 볼륨 생성 중..."
    docker volume create allora_monitor_data
fi

# 볼륨에 적절한 권한 설정
echo "데이터 볼륨 권한 설정 중..."
docker run --rm -v allora_monitor_data:/data alpine sh -c "chmod -R 777 /data"

# docker-compose.yml 파일을 수정하여 스웜 모드에서 지원되지 않는 옵션 제거
echo "Docker Compose 파일 준비 중..."
cp docker-compose.yml docker-compose.swarm.yml

# container_name 항목 제거
sed -i '' -e 's/container_name:.*//g' docker-compose.swarm.yml 2>/dev/null || \
sed -i 's/container_name:.*//g' docker-compose.swarm.yml

# 스택 배포
echo "스택 배포 중..."
docker stack deploy -c docker-compose.swarm.yml allora-monitor

# 임시 파일 정리
rm -f docker-compose.swarm.yml

# 서비스가 모두 시작될 때까지 대기
echo "서비스 배포 대기 중..."
for service in $(docker service ls --filter "name=allora-monitor" --format "{{.Name}}"); do
    echo "서비스 $service 배포 대기 중..."
    
    while true; do
        # 실행 중인 레플리카 수와 원하는 레플리카 수를 가져옴
        REPLICAS=$(docker service ls --filter "name=$service" --format "{{.Replicas}}")
        CURRENT=$(echo $REPLICAS | cut -d'/' -f1)
        DESIRED=$(echo $REPLICAS | cut -d'/' -f2)
        
        if [ "$CURRENT" = "$DESIRED" ]; then
            echo "서비스 $service가 성공적으로 배포되었습니다 ($CURRENT/$DESIRED)"
            break
        fi
        
        echo "서비스 $service: $CURRENT/$DESIRED 레플리카 실행 중..."
        sleep 3
    done
done

# 배포 확인
echo -e "\n배포된 서비스 확인:"
docker service ls | grep allora-monitor

echo -e "\n프론트엔드 레플리카 확인:"
docker service ps allora-monitor_frontend

echo -e "\n배포가 완료되었습니다."
echo "애플리케이션은 http://localhost:280 또는 설정된 도메인으로 접근할 수 있습니다." 