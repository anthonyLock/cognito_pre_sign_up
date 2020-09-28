PHONY: build deploy
PROFILE?= 

get:
	go get ./...

test:
	go test  ./...

build:
	rm -rf bin
	env GOOS=linux go build -ldflags="-s -w" -o bin/presignup main.go

clean:
	rm -rf ./bin

deploy: clean build test 
	env AWS_PROFILE=${PROFILE} sls deploy