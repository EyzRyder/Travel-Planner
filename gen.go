package main

//go:generate goapi-gen --package=spec --out ./internal/api/spec/journey.spec.go ./internal/api/spec/journey.spec.json run go generate ./... | run go generate
//go:generate tern migrate --migrations ./internal/pgstore/migrations --config ./internal/pgstore/migrations/tern.conf run go generate ./... | run go generate
//go:generate sqlc generate -f ./internal/pgstore/sqlc.yml
