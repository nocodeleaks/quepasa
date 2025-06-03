# Build stage
FROM golang:latest as builder

WORKDIR /build

COPY /docker/docker-entrypoint.sh .
COPY /src/ .

RUN chmod 755 docker-entrypoint.sh
RUN sed -i 's/\r//' docker-entrypoint.sh
RUN go build -o service main.go

# Run stage
FROM golang:latest as src

# 1) install ffmpeg here
RUN apt-get update
RUN apt-get install -y --no-install-recommends ffmpeg
RUN rm -rf /var/lib/apt/lists/*

WORKDIR /opt/quepasa/

COPY --from=builder /build/. /builder/
COPY --from=builder /build/docker-entrypoint.sh /

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["sh"]
