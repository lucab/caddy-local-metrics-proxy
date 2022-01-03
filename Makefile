caddy:
	xcaddy build --with github.com/lucab/caddy-local-metrics-proxy=./

build: caddy

all: build

.PHONY: build all
