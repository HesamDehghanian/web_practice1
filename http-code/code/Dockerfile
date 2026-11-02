FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/proxy .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/proxy /proxy
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/proxy"]
