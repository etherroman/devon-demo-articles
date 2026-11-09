FROM golang:1.25.7-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/articles ./cmd/server \
    && CGO_ENABLED=0 go build -trimpath -o /out/manage ./cmd/manage

FROM alpine:3.23
RUN addgroup -g 10001 articles && adduser -D -u 10001 -G articles articles \
    && mkdir -p /app/data && chown -R articles:articles /app
WORKDIR /app
COPY --from=build /out/ /usr/local/bin/
USER articles
ENV HTTP_ADDR=0.0.0.0:8082 DB_PATH=/app/data/articles.db GIN_MODE=release
EXPOSE 8082
CMD ["articles"]
