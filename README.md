# near-drpc
## A Decentralized RPC service for NEAR Protocol 

The service will scan the network and try to find out all the RPC nodes.
Sort the RPC nodes and establish a DRPC proxy server with these nodes which have the lowest latency.

Nodes are auto reweighted by latency and kicked out by status if not synced up.

`"Keep your service always online"`, that is what DRPC does.


# Features
- [x] Dashboard: a web UI for the service
- [x] Upstream: dispatch http jsonrpc to http backends
- [x] Expose http jsonrpc interfaces
- [x] Load-balance: WeightedRound-Robin algorithm
- [ ] Statistic: calculate metrics, query and display on the dashboard
- [ ] Upstream: dispatch http or websocket jsonrpc to http or websocket backends
- [ ] Cache: cache responses to some methods
- [ ] Offline requests: serving by local block data storage


# Usage
## Getting Started

### Building from source
- Install [Go](https://golang.org/doc/install)
```sh
git clone https://github.com/blockpilabs/near-drpc.git
go build
```

### Docker
```sh
docker run -it -p 8181:8181 -p 9191:9191 blockpi/near-drpc
```

<div style="page-break-after: always;"></div>

## DRPC status api
```sh
http://localhost:8181/api/status
```

## DRPC proxy endpoint
```sh
http://localhost:9191/
```
The backend upstream is set in the response headers.
```sh
curl \
  --data '{"method":"status","params":[],"id":1,"jsonrpc":"2.0"}' \
  -H "Content-Type: application/json" \
  -X POST http://localhost:9191/ \
  -D -

HTTP/1.1 200 OK
Access-Control-Allow-Methods: GET, POST, PATCH, PUT, DELETE, OPTIONS
Access-Control-Allow-Origin: *
Content-Type: application/json
Upstream: http://34.81.xxx.xxx:3030
Date: Sat, 18 Sep 2021 03:28:37 GMT
Transfer-Encoding: chunked

{"id":1,"jsonrpc":"2.0","result":...}
```
