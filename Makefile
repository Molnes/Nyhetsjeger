build:
	templ generate -path ./api/web/
	npx tailwindcss -o assets/css/tailwind.css
	go build -o bin/main cmd/main.go

run:
	templ generate -path ./api/web/
	npx tailwindcss -o assets/css/tailwind.css
	go run cmd/main.go

live_reload:
	air -c .air.toml
