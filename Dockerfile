FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

RUN templ generate
RUN go build -o ./bin/main ./cmd/server/main.go

FROM node:latest

WORKDIR /app

COPY --from=builder /app .

RUN npx --yes tailwindcss build -i assets/css/styles.css -o assets/css/tailwind.css

EXPOSE 8080

CMD ["./bin/main"]
