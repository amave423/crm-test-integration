FROM golang:1.25-alpine AS build

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /bin/test-constructor ./cmd/test-constructor

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/test-constructor /app/test-constructor

EXPOSE 8080
CMD ["/app/test-constructor"]
