build:
	templ generate
	npx tailwindcss -o assets/css/tailwind.css
	go build -o bin/main cmd/main.go

run:
	go run cmd/main.go

live_reload:
	air -c .air.toml
