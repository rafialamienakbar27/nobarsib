# NOBARSIB — perintah pengembangan
#
# Jalankan `make` tanpa argumen untuk melihat daftar target.

API_DIR        := nobarsib-api
WEB_DIR        := nobarsib-web
ANDROID_DIR    := nobarsib-android
MIGRATIONS_DIR := $(API_DIR)/migrations
SEED_DIR       := $(API_DIR)/seed

# Muat .env kalau ada, supaya DATABASE_URL tersedia untuk migrate & psql.
-include $(API_DIR)/.env
export

# golang-migrate dijalankan lewat `go run` supaya tidak perlu instalasi terpisah.
MIGRATE := go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

# Docker Compose v2 memakai subcommand `docker compose`; instalasi lama memakai
# biner terpisah `docker-compose`. Pilih yang tersedia.
COMPOSE := $(shell if docker compose version >/dev/null 2>&1; \
	then echo "docker compose"; else echo "docker-compose"; fi)

.DEFAULT_GOAL := help

## help: tampilkan daftar target
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

# ---------------------------------------------------------------- lingkungan

## up: nyalakan db, redis, api, worker
up:
	$(COMPOSE) up -d --build

## down: matikan semua service (data tetap tersimpan)
down:
	$(COMPOSE) down

## reset: matikan service DAN hapus volume — seluruh data hilang
reset:
	$(COMPOSE) down -v

## logs: ikuti log semua service
logs:
	$(COMPOSE) logs -f

## infra: nyalakan db dan redis saja, untuk `make run` di host
infra:
	$(COMPOSE) up -d db redis

## ps: status service
ps:
	$(COMPOSE) ps

# ------------------------------------------------------------------ migrasi

## migrate-up: jalankan semua migrasi yang belum diterapkan
migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

## migrate-down: mundurkan satu migrasi
migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

## migrate-reset: mundurkan seluruh migrasi sampai nol
migrate-reset:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all

## migrate-version: versi migrasi yang sedang aktif
migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

## migrate-force v=N: bersihkan status dirty dengan memaksa versi ke N
migrate-force:
	@test -n "$(v)" || (echo "pakai: make migrate-force v=0" && exit 1)
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(v)

## migrate-create name=xxx: buat pasangan file migrasi baru
migrate-create:
	@test -n "$(name)" || (echo "pakai: make migrate-create name=nama_migrasi" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## seed: isi data referensi (competition, team) — aman diulang
seed:
	@for f in $(SEED_DIR)/*.sql; do \
		echo "seed: $$f"; \
		$(COMPOSE) exec -T db psql -v ON_ERROR_STOP=1 -U nobarsib -d nobarsib < $$f || exit 1; \
	done

## fixture: isi data uji pengembangan (MENGHAPUS venue, event, dan match yang ada)
fixture:
	$(COMPOSE) exec -T db psql -v ON_ERROR_STOP=1 -U nobarsib -d nobarsib < $(API_DIR)/testdata/dev_fixture.sql

## psql: buka shell psql di container database
psql:
	$(COMPOSE) exec db psql -U nobarsib -d nobarsib

# --------------------------------------------------------------------- kode

## run: jalankan API di host (butuh `make infra` lebih dulu)
run:
	cd $(API_DIR) && go run ./cmd/api

## run-worker: jalankan worker di host
run-worker:
	cd $(API_DIR) && go run ./cmd/worker

## admin-create EMAIL=.. NAME=..: buat akun admin (kata sandi ditanya lewat prompt)
admin-create:
	@test -n "$(EMAIL)" || (echo 'pakai: make admin-create EMAIL=kamu@contoh.id NAME="Nama Kamu"' && exit 1)
	cd $(API_DIR) && go run ./cmd/adminctl create "$(EMAIL)" "$(NAME)"

## import FILE=..: impor venue dari berkas JSON (idempoten berdasarkan slug)
import:
	@test -n "$(FILE)" || (echo 'pakai: make import FILE=nobarsib-api/testdata/venues.json' && exit 1)
	cd $(API_DIR) && go run ./cmd/importer -file "$(CURDIR)/$(FILE)"

## import-check FILE=..: periksa berkas venue tanpa menulis ke database
import-check:
	@test -n "$(FILE)" || (echo 'pakai: make import-check FILE=nobarsib-api/testdata/venues.json' && exit 1)
	cd $(API_DIR) && go run ./cmd/importer -file "$(CURDIR)/$(FILE)" -dry-run

# ----------------------------------------------------------------- frontend

## web: jalankan frontend mode pengembangan di http://localhost:3000
web:
	cd $(WEB_DIR) && npm run dev

## web-build: build produksi frontend
web-build:
	cd $(WEB_DIR) && npm run build

## web-start: jalankan hasil build produksi frontend
web-start:
	cd $(WEB_DIR) && npm start

## web-check: type-check + lint frontend
web-check:
	cd $(WEB_DIR) && npx tsc --noEmit && npx eslint .

## web-icons: buat ulang ikon PWA
web-icons:
	cd $(WEB_DIR) && node scripts/generate-icons.mjs

## apk: bangun ulang APK Android dan pasang ke nobarsib-web/public/
apk:
	cd $(ANDROID_DIR) && ./build.sh $(VERSI)

## build: kompilasi kedua biner ke nobarsib-api/bin/
build:
	cd $(API_DIR) && go build -o bin/api ./cmd/api && go build -o bin/worker ./cmd/worker

## test: jalankan test dengan race detector dan coverage
test:
	cd $(API_DIR) && go test ./... -race -cover

## vet: analisis statis bawaan Go
vet:
	cd $(API_DIR) && go vet ./...

## fmt: rapikan format kode
fmt:
	cd $(API_DIR) && go fmt ./...

## tidy: rapikan go.mod dan go.sum
tidy:
	cd $(API_DIR) && go mod tidy

## check: pemeriksaan backend + frontend, dijalankan sebelum commit
check: fmt vet test web-check

# ------------------------------------------------------------------ bantuan

## health: panggil endpoint /health
health:
	@curl -sS http://localhost:8080/health | (command -v jq >/dev/null && jq || cat)

.PHONY: help up down reset logs infra ps \
        migrate-up migrate-down migrate-reset migrate-version migrate-force migrate-create \
        seed fixture psql run run-worker build test vet fmt tidy check health \
        admin-create import import-check \
        web web-build web-start web-check web-icons
