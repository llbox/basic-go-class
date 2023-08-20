.PHONY: docker
docker:
	@del webook
	@set GOOS=linux
	@set GOARCH=amd64
	@go build -o webook main.go
	@docker rmi ruler/webook:v0.0.1
	@docker build -t ruler/webook:v0.0.1 .