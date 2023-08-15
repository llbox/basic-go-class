.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux GOARCH=arm go build -o webook .
	@docker rmi ruler/webook:v0.0.1
	@docker build -t ruler/webook:v0.0.1 .