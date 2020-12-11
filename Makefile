test:
	go test -short -v ./...

storetest:
	docker-compose -f deployments/storetest.yml up --build --abort-on-container-exit
	docker-compose down

.PHONY: test storetest
