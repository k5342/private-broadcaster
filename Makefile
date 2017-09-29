init:
	GOOS=linux GOARCH=amd64 go build app.go && \
	docker build -t private-broadcaster:0.5 .
