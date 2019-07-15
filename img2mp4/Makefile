docker:
	GOOS=linux GOARCH=amd64 go build -o bin/server
	docker build -t img2mp4 .

run:
	docker run --rm -it -p 1234:1234 img2mp4
