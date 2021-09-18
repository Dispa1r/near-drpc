package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/blockpilabs/near-drpc/dashboard"
	"github.com/blockpilabs/near-drpc/log"
	"github.com/blockpilabs/near-drpc/near"
	"github.com/blockpilabs/near-drpc/providers"
	"github.com/blockpilabs/near-drpc/proxy"
	"github.com/valyala/fasthttp"
)

var(
	logger = log.GetLogger("main")

	Peers map[string]string

	Producers map[string]string
)

func init() {
	Peers = make(map[string]string)
	Producers = make(map[string]string)
}

func main() {
	seed := "https://public-rpc.blockpi.io/http/near"

	provider := providers.NewHttpJsonRpcProvider(":9191", "/", &providers.HttpJsonRpcProviderOptions{
		TimeoutSeconds: 30,
	})
	server := proxy.NewProxyServer(provider)

	nearDRpc := near.NewNearDRpcService(server, seed)
	nearDRpc.Run()

	go server.Start()

	dashboard.ListenAndServ(dashboard.NewRouter(nearDRpc))
}




func main2() {
	//https://public-rpc.blockpi.io/http/near
	Peers["https://public-rpc.blockpi.io/http/near"] = "https://public-rpc.blockpi.io/http/near"
	//Peers["https://rpc.mainnet.near.org"] = "https://rpc.mainnet.near.org"
	//Peers["http://135.181.118.199:3030"] = "http://135.181.118.199:3030"

	for true {
		for _, peer := range Peers{
			getPeers(peer)
			/*
			rpcs := getPeers(peer)
			if rpcs != nil {
				for _, rpc := range *rpcs {
					getPeers(rpc)
				}
			}
			*/

		}

		time.Sleep(time.Second/2)
	}

	fmt.Println("")
}


func getPeers(rpcEndPoint string) *map[string]string {
	defer func() {
		if err:=recover(); err!=nil {
			fmt.Println("UDPRecv error: ", err)
		}
	}()

	req := &fasthttp.Request{}
	req.SetRequestURI(rpcEndPoint)
	requestBody := []byte(`{"jsonrpc":"2.0","id":1, "method":"network_info","params": []}`)
	req.SetBody(requestBody)
	req.Header.SetContentType("application/json")
	req.Header.SetMethod("POST")

	resp := &fasthttp.Response{}
	client := &fasthttp.Client{}
	if err := client.DoTimeout(req, resp, time.Second * 10);err != nil {
		//fmt.Println(rpcEndPoint, 0, "peers")
		return nil
	}
	//respStr := string(resp.Body())
	//fmt.Println(respStr)

	var rpcs = make(map[string]string)

	var result map[string]interface{}
	_ = json.Unmarshal(resp.Body(), &result)
	peers := result["result"].(map[string]interface{})["active_peers"].([]interface{})
	for _, peer := range peers {
		addr := peer.(map[string]interface{})["addr"].(string)
		ip := strings.Split(addr,":")[0]
		rpcs[ip] = "http://"+ip + ":3030"
		id :=  peer.(map[string]interface{})["id"].(string)
		if va, ok := Producers[id]; ok {
			fmt.Println("Find", va, addr)
		}

		//if id == "ed25519:5pRiNCHbXFALoVqHJtqvUdJadT85r6L3GVSnyUeS3GJX" {
		//	fmt.Println("Find hashquark node:", addr)
		//}
		//if id == "ed25519:83z1y1SL1CNCVWJiqukTKTVng9by7SGkvnTT8THQh8TS" {
		//	fmt.Println("Find buildlinks node:", addr)
		//}
	}
	producers := result["result"].(map[string]interface{})["known_producers"].([]interface{})
	for _, producer := range producers {
		va := producer.(map[string]interface{})["account_id"].(string)
		peerId :=  producer.(map[string]interface{})["peer_id"].(string)
		Producers[peerId] = va
	}

	if len(rpcs) > 0 {
		if _,ok:=Peers[rpcEndPoint]; !ok {
			fmt.Println(rpcEndPoint, len(rpcs) , "peers")
			Peers[rpcEndPoint] = rpcEndPoint
		}
	}


	return &rpcs
}


