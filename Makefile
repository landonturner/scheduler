run:
	env JWT_SECRET=70244c6405d1ef34146f724bae668c5673068b220f79e34540a7b25e42c1a3fa \
	go run server.go

test:
	go test ./...

build:
	docker build . -t scheduler
