FROM golang:1.11 AS builder
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY *.go ./
CMD /bin/sh
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /app .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
COPY --from=builder /app /app
CMD ["/app"]
