docker-build:
	docker build . -t hermeznet/hermez-node:latest

build:
	make docker-build
	docker-compose up