run:
	go run cmd/main.go

start:
	docker compose up --build

gen-crt:
	scripts/gen_crt.sh