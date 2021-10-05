.PHONY: start-server start-client start-dummy-service \
	start-benchmark start-tunnel-benchmark start-baseline-benchmark

# This invocation starts a server that listens on port 9001 and 9002
# any connection to localhost:9002 will be tunneled via a websocket connection
# on 9001.
#
# Each websocket connection will proxy a single connection to localhost:9002
start-server: dist
	./dist/kubetunnel server -bind localhost:9001 -target localhost:9002

# This invocation will connect to the server at localhost:9001 and for every
# connection received in the websocket it will proxy the connection to
# localhost:8081.
#
# Effectively it is exposing localhost:8081 on the port desigined by the server
# running at localhost:9001
start-client: dist
	./dist/kubetunnel client -server ws://localhost:9001/ws/anything -local localhost:8081

# This invocation will start a simple websocket bus
start-websocket-bus: dist
	./dist/kubetunnel websocket-bus -bind localhost:8082 -directory $(PWD)/internal/demo/wsbus/site

start-dummy-service: dist
	./dist/kubetunnel static-file-server -bind localhost:8081 -directory $(PWD)

vegetaDuration?=30s
vegetaParameters?=-duration=$(vegetaDuration) -max-connections 100 -workers 50 -rate 100

# start-benchmark will runn the baseline benchmark and right after
# it will run the vegeta benchmark
start-benchmark:
	$(MAKE) -C . start-baseline-benchmark vegetaParameters="$(vegetaParameters)" || true
	$(MAKE) -C . start-tunnel-benchmark vegetaParameters="$(vegetaParameters)" || true

# start-tunnel-benchmark uses the vegeta load test tool (https://github.com/tsenart/vegeta)
# and you must have started the following tasks on a different terminal/session
#
# make start-server
#
# make start-client
#
# make start-dummy-service
start-tunnel-benchmark: dist
	echo "GET http://localhost:9002/" | \
		vegeta attack $(vegetaParameters)\
			| tee ./dist/result-tunnel.bin \
			| vegeta plot > ./dist/tunnel.html; \
				vegeta report < ./dist/result-tunnel.bin

# start-baseline-benchmark will run the same benchmark as the start-tunnel-benchmark but instead of calling
# the tunnel endpoint it will call the dummy-service directly, this is used to compare
# the overhead introduced by the tunnel in the whole process
start-baseline-benchmark: dist
	echo "GET http://localhost:8081/" | \
		vegeta attack $(vegetaParameters)\
			| tee ./dist/result-baseline.bin \
			| vegeta plot > ./dist/baseline.html; \
				vegeta report < ./dist/result-baseline.bin
