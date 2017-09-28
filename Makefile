init:
	cd src && \
	GOOS=linux GOARCH=amd64 go build app.go && \
	cd .. && \
	docker build -t private-broadcaster:0.5 .
