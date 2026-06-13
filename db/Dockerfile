FROM postgres:latest

# Copy initialization scripts
# Scripts in /docker-entrypoint-initdb.d/ are automatically run during container startup
COPY *.sql /docker-entrypoint-initdb.d/
