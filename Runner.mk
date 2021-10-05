.PHONY: start-server start-client

start-server: dist
	./dist/kubetunnel server -bind localhost:9001

start-client: dist
	./dist/kubetunnel client -server localhost:9001 -local localhost:8081 -remote localhost:9002
