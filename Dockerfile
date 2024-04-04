FROM node:21 AS tailwind-builder
WORKDIR /app

COPY assets/css/styles.css ./assets/css/
COPY internal/web_server/web/views ./internal/web_server/web/views/
COPY tailwind.config.js ./

RUN npm install tailwindcss@latest
RUN npx tailwindcss build -i assets/css/styles.css -o assets/css/tailwind.css

FROM golang:1.22 AS go-builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN templ generate
RUN go build -o ./bin/main ./cmd/server/main.go

FROM golang:1.22
WORKDIR /app

COPY --from=go-builder /app/bin ./
COPY assets/ ./assets/
COPY --from=tailwind-builder /app/assets/css/tailwind.css ./assets/css/tailwind.css

EXPOSE 8080

CMD ["./main"]