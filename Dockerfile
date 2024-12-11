FROM golang:1.23.3-alpine
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod tidy
COPY . .
RUN go build -o server cmd/server/main.go
RUN chmod +x server
EXPOSE 4000
CMD [ "./server" ]