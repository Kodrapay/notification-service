FROM alpine:latest

# Install curl for health checks
RUN apk --no-cache add curl

WORKDIR /root/

# Copy pre-built binary (context is the service directory)
COPY notification-service .

EXPOSE 7000

CMD ["./notification-service"]
