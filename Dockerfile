FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hospital-tracker .

FROM gcr.io/distroless/static-debian12
WORKDIR /
COPY --from=builder /app/hospital-tracker .
COPY --from=builder /app/.env.example .env
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["./hospital-tracker"]