FROM golang:1.22.1 AS fetch
WORKDIR /app
COPY go.mod go.sum
RUN go mod download

FROM ghcr.io/a-h/templ:latest AS generate
WORKDIR /app
COPY --chown=65532:65532 . .
RUN ["templ", "generate"]

FROM golang:1.22.1 as build
WORKDIR /app
COPY --from=generate /app .
RUN CGO_ENABLED=0 GOOS=linux go build -buildvcs=false -o app

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/app .
EXPOSE 8080
CMD [ "/app/app" ]
