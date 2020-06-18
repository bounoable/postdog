test:
	go test -short ./...

test-full:
	docker-compose -f deployments/testing.compose.yml up --build --abort-on-container-exit
	docker-compose -f deployments/testing.compose.yml down

.PHONY: test test-full
