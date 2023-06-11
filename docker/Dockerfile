# Build stage
FROM golang:latest as builder
WORKDIR /build
ENV MIGRATIONS=./migrations
COPY /src/ .
RUN go build -o service main.go

# Run stage
FROM golang:latest as src
WORKDIR /opt/quepasa/
ENV MIGRATIONS=./migrations
COPY --from=builder /build /opt/quepasa
EXPOSE 31000
ENTRYPOINT ["/opt/quepasa/service"]