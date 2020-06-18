test:
	go test -short ./...

test-full:
	make docker-up
	MONGO_URI=mongodb://mongo:27017 \
		go test -v ./...
	make docker-down

docker-up:
	docker-compose -f deployments/testing.yml up -d

docker-down:
	docker-compose -f deployments/testing.yml down

.PHONY: test test-full docker-up docker-down
