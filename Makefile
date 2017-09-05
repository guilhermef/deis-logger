build:
	@mkdir -p bin && go build -o ./bin/deis-logger main.go

build-docker:
	@docker build -t deis-logger .

cross-build-linux-amd64:
	@env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/deis-logger-linux-amd64
	@chmod a+x ./bin/deis-logger-linux-amd64
