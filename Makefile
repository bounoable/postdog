test:
	go test -short ./...

test-full:
	docker-compose -f deployments/testing.yml up --build --abort-on-container-exit
	docker-compose -f deployments/testing.yml down

.PHONY: test test-full
