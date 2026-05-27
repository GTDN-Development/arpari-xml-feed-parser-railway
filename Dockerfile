FROM golang:1.26-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o /out/server ./cmd/server \
    && go build -o /out/rebuild ./cmd/rebuild

FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /out/server /app/server
COPY --from=build /out/rebuild /app/rebuild

EXPOSE 8080

CMD ["/app/server"]
