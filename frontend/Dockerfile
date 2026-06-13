FROM caddy:2-alpine

# Copy the configuration file
COPY Caddyfile /etc/caddy/Caddyfile

# Copy the static assets
COPY index.html /srv/index.html
COPY assets /srv/assets
COPY pages /srv/pages

# Expose port 80 (Caddy default)
EXPOSE 80
