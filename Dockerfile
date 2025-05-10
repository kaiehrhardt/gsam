# hadolint global ignore=DL3007
FROM ghcr.io/a-h/templ:latest AS generate
WORKDIR /app
COPY --chown=65532:65532 main.templ .
RUN ["templ", "generate"]

FROM registry.hub.docker.com/library/golang:1.24.3 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY --from=generate /app/main_templ.go .
COPY *.go .
RUN go build -o app

FROM gcr.io/distroless/base:latest
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
USER nonroot:nonroot
CMD [ "/app/app" ]
