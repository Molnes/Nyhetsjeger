FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

Run go install github.com/a-h/templ/cmd/templ@latest
Run npx tailwindcss -o assets/css/tailwind.css

COPY . .

RUN make build

CMD ["./bin/main"]
