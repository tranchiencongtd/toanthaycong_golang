# Giai đoạn build
FROM golang:1.19-alpine AS builder

# Thiết lập thư mục làm việc
WORKDIR /app

# Copy các file go mod
COPY go.mod go.sum ./

# Tải xuống dependencies
RUN go mod download

# Copy source code
COPY . .

# Build ứng dụng
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/_your_app_/

# Giai đoạn runtime
FROM alpine:latest

# Cài đặt ca-certificates cho HTTPS requests
RUN apk --no-cache add ca-certificates

# Tạo thư mục app
WORKDIR /root/

# Copy binary từ giai đoạn builder
COPY --from=builder /app/main .

# Copy config files nếu cần
COPY --from=builder /app/configs ./configs

# Mở port
EXPOSE 8080

# Chạy ứng dụng
CMD ["./main"]
