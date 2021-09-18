package near

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/blockpilabs/near-drpc/network/http"
	"github.com/blockpilabs/near-drpc/rpc"
	"github.com/blockpilabs/near-drpc/utils"
)

const DEFAULT_RPC_PORT = 3030

type Peer struct {
	Id    		string
	RpcAddr  	string
	LastOnline 	int64
	Latency		int

	Height		int64
	Syncing		bool

}

type Peers []*Peer

func CreateDefaultRpcAddress(uri string) string {
	ip := net.ParseIP(uri)
	if ip == nil {
		return uri
	}
	return fmt.Sprintf("http://%s:%d", ip, DEFAULT_RPC_PORT)
}

func (peer *Peer)Peers() Peers{
	startTime := utils.CurrentTimestampMilli()
	//peers := GetPeers(peer.RpcAddr)
	jsonReq := "{\"jsonrpc\":\"2.0\",\"id\":1, \"method\":\"network_info\",\"params\": []}"
	if data := http.PostJson(peer.RpcAddr, jsonReq); data != nil {
		resp := rpc.JSONRpcResponse{}
		if err := json.Unmarshal(data, &resp); err == nil && resp.Error == nil {
			activePeers := resp.Result.(map[string]interface{})["active_peers"].([]interface{})

			peers := make([]*Peer, len(activePeers))
			for idx, peer := range activePeers {
				ip := strings.Split(peer.(map[string]interface{})["addr"].(string), ":")[0]
				id :=  peer.(map[string]interface{})["id"].(string)
				peers[idx] = NewPeer(ip, id)
			}

			endTime := utils.CurrentTimestampMilli()
			peer.Latency = int(endTime - startTime)

			peer.LastOnline = endTime

			return peers
		}
	}
	return nil
}

func (peer *Peer)UpdateStatus(){
	startTime := utils.CurrentTimestampMilli()
	jsonReq := "{\"jsonrpc\":\"2.0\",\"id\":1, \"method\":\"status\",\"params\": []}"
	if data := http.PostJson(peer.RpcAddr, jsonReq); data != nil {
		resp := rpc.JSONRpcResponse{}
		if err := json.Unmarshal(data, &resp); err == nil && resp.Error == nil {
			syncInfo := resp.Result.(map[string]interface{})["sync_info"].(map[string]interface{})
			peer.Syncing = syncInfo["syncing"].(bool)
			peer.Height = int64(syncInfo["latest_block_height"].(float64))

			endTime := utils.CurrentTimestampMilli()
			peer.Latency = int(endTime - startTime)
			peer.LastOnline = endTime
		}
	}
}



func (peer *Peer)IsOnline() bool{
	return utils.CurrentTimestampMilli() - peer.LastOnline < 300 * 1000
}


func NewPeer(uri string, id string) *Peer {
	return &Peer{
		Id: id,
		RpcAddr: CreateDefaultRpcAddress(uri),

		Syncing: true,
	}
}

func (peers Peers)Len() int {
	return len(peers)
}

func (peers Peers)Less(i, j int) bool {
	return peers[i].Latency <  peers[j].Latency
}

func (peers Peers)Swap(i, j int) {
	tmp := peers[i]
	peers[i] = peers[j]
	peers[j] = tmp
}

func (peers Peers)String() string{
	str := ""
	for _ , peer := range peers{
		str += "  -  " + fmt.Sprintf("%s %dms", peer.RpcAddr, peer.Latency)
	}
	return  str
}