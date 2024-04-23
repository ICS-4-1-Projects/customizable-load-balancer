FROM golang:1.22.2-alpine as builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o main .


FROM alpine:latest  
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 5000
CMD ["./main"]
