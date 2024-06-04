# hadolint global ignore=DL3007
FROM golang:1.22.4 AS fetch
WORKDIR /app
COPY go.mod go.sum
RUN go mod download

FROM ghcr.io/a-h/templ:latest AS generate
WORKDIR /app
COPY --chown=65532:65532 . .
RUN ["templ", "generate"]

FROM golang:1.22.4 as build
WORKDIR /app
COPY --from=generate /app .
RUN GOOS=linux go build -buildvcs=false -o app

FROM gcr.io/distroless/base:latest
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
CMD [ "/app/app" ]
