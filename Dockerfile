FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/httpserver ./cmd/httpserver
RUN CGO_ENABLED=0 go build -o /out/migrate ./cmd/migrate

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/httpserver /app/httpserver
COPY --from=build /out/migrate /app/migrate
COPY sql ./sql
EXPOSE 8080
CMD ["/app/httpserver"]
