run:
	go run cmd/main.go

.PHONY: gen-crt
gen-crt:
	scripts/gen_crt.sh

.PHONY: add-cert
add-cert:
	sudo apt-get install -y ca-certificates | sudo cp ca.crt /usr/local/share/ca-certificates |  sudo update-ca-certificates

.PHONY: up-compose
up-compose:
	docker compose up --build

.PHONY: all-cert
all-cert: gen-crt add-cert

start: gen-crt add-cert up-compose