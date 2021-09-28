run:
	air

mod:
	go mod vendor

build:
	go build -mod=vendor

init:
	go get github.com/gorilla/sessions
	go mod vendor
	go mod tidy

gcrbuild:
	docker build -f Dockerfile --tag=asia.gcr.io/work-317400/sample:latest .
