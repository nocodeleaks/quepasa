# Frontend build stage
FROM node:20-alpine as frontend_builder

WORKDIR /frontend

COPY /src/apps/vuejs/package*.json ./
RUN npm ci

COPY /src/apps/vuejs/ ./
RUN npm run build

# Build stage
FROM golang:latest as builder

WORKDIR /build

COPY /docker/docker-entrypoint.sh .
COPY /src/ .
COPY --from=frontend_builder /frontend/dist ./apps/vuejs/dist

RUN chmod 755 docker-entrypoint.sh
RUN sed -i 's/\r//' docker-entrypoint.sh
RUN go build -o service main.go

# Run stage
FROM golang:latest as src

# Install ffmpeg (video/audio) and poppler-utils (PDF thumbnails)
RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg poppler-utils && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /opt/quepasa/

COPY --from=builder /build/. /builder/
COPY --from=builder /build/docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["sh"]
