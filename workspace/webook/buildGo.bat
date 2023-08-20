SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -tags=k8s -o webook main.go