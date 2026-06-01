FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /out/server ./cmd/server \
    && go build -o /out/rebuild ./cmd/rebuild \
    && go build -o /out/rebuild-trigger ./cmd/rebuild-trigger

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /out/rebuild /app/rebuild
COPY --from=build /out/rebuild-trigger /app/rebuild-trigger

EXPOSE 8080

CMD ["/app/server"]
