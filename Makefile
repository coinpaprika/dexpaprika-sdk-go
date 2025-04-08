.PHONY: run-basic run-comprehensive run-example test

run-basic:
	go run cmd/basic/main.go

run-comprehensive:
	go run cmd/comprehensive/main.go

run-example:
	go run examples/production_usage.go

test:
	go test ./dexpaprika/... -v 