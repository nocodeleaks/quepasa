# Build stage
FROM golang:latest as builder
WORKDIR /build
COPY /src/ .
RUN go build -o service main.go

# Run stage
FROM golang:latest as src

ENV WEBSOCKETSSL=false
ENV WEBAPIPORT=31000
ENV APP_ENV=production
ENV MIGRATIONS=/opt/quepasa/migrations
ENV DEBUGJSONMESSAGES=false
ENV HTTPLOGS=false

WORKDIR /opt/quepasa/
COPY --from=builder /build /opt/quepasa
EXPOSE 31000

ENTRYPOINT ["/opt/quepasa/service"]