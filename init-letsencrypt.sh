#!/bin/bash

if ! [ -x "$(command -v docker-compose)" ]; then
  echo 'Error: docker-compose is not installed.' >&2
  exit 1
fi

domains=(allora-inference.kro.kr backend.allora-inference.kro.kr)
rsa_key_size=4096
data_path="./nginx/certbot"
email="" # Let's Encrypt 이메일 주소 (선택 사항)

# Nginx 설정 파일이 존재하는지 확인
if [ ! -e "$data_path/conf/options-ssl-nginx.conf" ] || [ ! -e "$data_path/conf/ssl-dhparams.pem" ]; then
  echo "### Downloading recommended TLS parameters ..."
  mkdir -p "$data_path/conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$data_path/conf/options-ssl-nginx.conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$data_path/conf/ssl-dhparams.pem"
  echo
fi

# DH 파라미터 생성
echo "### Creating DH parameters for additional security ..."
mkdir -p "./nginx/dhparam"
openssl dhparam -out "./nginx/dhparam/dhparam.pem" 2048
echo

# 각 도메인에 대해 인증서 생성
for domain in "${domains[@]}"; do
  echo "### Creating dummy certificate for $domain ..."
  path="/etc/letsencrypt/live/$domain"
  mkdir -p "$data_path/conf/live/$domain"
  docker-compose run --rm --entrypoint "\
    openssl req -x509 -nodes -newkey rsa:$rsa_key_size -days 1\
      -keyout '$path/privkey.pem' \
      -out '$path/fullchain.pem' \
      -subj '/CN=localhost'" certbot
  echo
done

echo "### Starting nginx ..."
docker-compose up --force-recreate -d nginx
echo

# 인증서 발급
for domain in "${domains[@]}"; do
  echo "### Requesting Let's Encrypt certificate for $domain ..."
  
  # 이메일 옵션 설정
  email_arg=""
  if [ -n "$email" ]; then
    email_arg="--email $email"
  fi

  # Let's Encrypt 인증서 발급
  docker-compose run --rm --entrypoint "\
    certbot certonly --webroot -w /var/www/certbot \
      $email_arg \
      -d $domain \
      --rsa-key-size $rsa_key_size \
      --agree-tos \
      --force-renewal" certbot
  echo
done

echo "### Reloading nginx ..."
docker-compose exec nginx nginx -s reload 