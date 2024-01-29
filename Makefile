build:
	templ generate
	npx tailwindcss -o web/static/css/tailwind.css
	go build -o bin/main cmd/main.go

run:
	go run cmd/main.go

live_reload:
	air -c .air.toml
