FROM caddy:2-alpine

# 因為已經在 frontend 資料夾內了，直接複製當前目錄下的檔案即可
COPY ./Caddyfile /etc/caddy/Caddyfile
COPY ./index.html /srv/index.html
COPY ./assets /srv/assets
COPY ./pages /srv/pages