.PHONY: build deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/sftp-upload main.go

deploy: build
	sls deploy --verbose --stage staging

dev: build
	sls offline start --stage dev
