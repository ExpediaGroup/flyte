build:
	dep ensure
	go test ./... -tags="integration acceptance"
	go build

run: build run-mongo
	./flyte-api &

stop: stop-mongo
	killall flyte-api

run-mongo:
	docker run -dp 27017:27017 --name mongo mongo:latest

stop-mongo:
	docker rm -f mongo

docker-build:
	docker build --rm -t flyte:latest .

docker-run: docker-build run-mongo
	docker run -p 8080:8080 -e FLYTE_MGO_HOST=mongo -d --name flyte --link mongo:mongo flyte:latest

docker-stop: stop-mongo
	docker rm -f flyte