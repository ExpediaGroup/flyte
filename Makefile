build:
	dep ensure
	go test ./... -tags="integration acceptance"
	go build

run: build run-mongo
	nohup ./flyte &> flyte.out &
	echo "OK: flyte successfully started, output is in flyte.out file"

stop: stop-mongo
	killall flyte

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

test:
	dep ensure
	go test ./...
	go test ./... -tags="integration acceptance"