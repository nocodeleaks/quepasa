# Build stage
FROM golang:latest as builder
WORKDIR /build
COPY /src/ .
RUN go build -o service main.go

# Run stage
FROM golang:latest as src
WORKDIR /opt/quepasa/
COPY --from=builder /build /opt/quepasa
EXPOSE 31000
ENTRYPOINT ["/opt/quepasa/service"]