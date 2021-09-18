package near

import (
	"container/list"
	"sort"
	"sync"
	"time"

	"github.com/blockpilabs/near-drpc/log"
	"github.com/blockpilabs/near-drpc/proxy"
	"github.com/blockpilabs/near-drpc/utils"
)

var logger = log.GetLogger("near")

var (
	QUEUE_PEER = 1
	QUEUE_RPC = 2
)

type Near struct {
	locker sync.Mutex

	proxyServer *proxy.ProxyServer

	seedRpcEndpoint string

	peers map[string]*Peer
	rpcPeers map[string]*Peer

	peersCheckQueue *list.List
	rpcCheckQueue *list.List

	bestRpcNodes *list.List
	bestRpcNodesLatency int
}

func (near *Near)DiscoverSeedPeers() {
	seedPeer := NewPeer(near.seedRpcEndpoint, "")
	for true {
		peers := seedPeer.Peers()
		if peers != nil {
			for _ , peer := range peers{
				near.AddToPeerQueue(peer)
			}
			break
		}else{
			logger.Error("empty seed peers")
			time.Sleep(time.Second)
		}
	}
}

func (near *Near)AddToPeerQueue(peer *Peer) {
	near.locker.Lock()
	defer near.locker.Unlock()

	found := false
	for e := near.peersCheckQueue.Front(); e != nil; e = e.Next() {
		if peer.Id == e.Value.(*Peer).Id {
			found = true
		}
	}
	if !found {
		near.peersCheckQueue.PushBack(peer)
	}


	near.peers[peer.Id] = peer
}


func (near *Near)AddToRpcQueue(peer *Peer) {
	near.locker.Lock()
	defer near.locker.Unlock()

	found := false
	for e := near.rpcCheckQueue.Front(); e != nil; e = e.Next() {
		if peer.Id == e.Value.(*Peer).Id {
			found = true
		}
	}
	if !found {
		near.rpcCheckQueue.PushBack(peer)
	}

	near.rpcPeers[peer.Id] = peer
}

func (near *Near)RemoveFromRpcQueue(peer *Peer) {
	near.locker.Lock()
	defer near.locker.Unlock()

	if _, exists := near.rpcPeers[peer.Id]; exists {
		delete(near.rpcPeers, peer.Id)
	}

	for e := near.rpcCheckQueue.Front(); e != nil; e = e.Next() {
		if peer.Id == e.Value.(*Peer).Id {
			near.rpcCheckQueue.Remove(e)
			break
		}
	}
}


func (near *Near)GetPeerFromQueue(queueNum int) *Peer {
	near.locker.Lock()
	defer near.locker.Unlock()
	var queue *list.List = nil

	switch queueNum {
	case QUEUE_PEER:
		queue = near.peersCheckQueue
	case QUEUE_RPC:
		queue = near.rpcCheckQueue
	}
	if queue != nil {
		e := queue.Front()
		if e != nil {
			return queue.Remove(e).(*Peer)
		}
	}
	return nil
}

func (near *Near)CheckPeers()  {
	threads := 10
	for thd := 0; thd < threads; thd++ {
		go func() {
			for true {
				thisPeer := near.GetPeerFromQueue(QUEUE_PEER)
				if thisPeer == nil {
					//logger.Debug("empty peers check queue")
					time.Sleep(time.Second * 5)
					continue
				}

				check := (utils.CurrentTimestampMilli() - thisPeer.LastOnline) > 60 * 1000
				if !check {
					near.AddToPeerQueue(thisPeer)
					continue
				}

				newPeers := thisPeer.Peers()
				if newPeers != nil {
					near.AddToRpcQueue(thisPeer)
					for _, newPeer := range newPeers {
						near.AddToPeerQueue(newPeer)
					}
				}
				near.AddToPeerQueue(thisPeer)
			}
		}()
	}
}

func (near *Near)CheckRpcPeers()  {
	threads := 10
	for thd := 0; thd < threads; thd++ {
		go func() {
			for true {
				thisPeer := near.GetPeerFromQueue(QUEUE_RPC)
				if thisPeer == nil {
					//logger.Debug("empty rpc check queue")
					time.Sleep(time.Second * 5)
					continue
				}

				check := (utils.CurrentTimestampMilli() - thisPeer.LastOnline) > 60 * 1000 || thisPeer.Height == 0
				if !check {
					near.AddToRpcQueue(thisPeer)
					continue
				}

				thisPeer.UpdateStatus()

				if thisPeer.IsOnline() {
					near.AddToRpcQueue(thisPeer)
				}
			}
		}()
	}
}


func (near *Near)Run()  {
	near.DiscoverSeedPeers()
	near.CheckPeers()
	near.CheckRpcPeers()

	go func() {
		for true {
			near.UpdateBestNodeByLatency()
			near.SetBestNodeToProxyServerBackends()
			time.Sleep(time.Second * 30)
		}
	}()

	go func() {
		for true {
			time.Sleep(time.Second * 10)
			logger.Info("Peer: ", len(near.peers), ", Rpc Node: ", len(near.rpcPeers),
				", Queue: ", near.peersCheckQueue.Len(), ", ", near.rpcCheckQueue.Len())
		}
	}()

}

func NewNearDRpcService(proxyServer *proxy.ProxyServer, seedRpcEndpoint string) *Near{
	return &Near{
		proxyServer: proxyServer,
		seedRpcEndpoint: seedRpcEndpoint,
		peers: make(map[string]*Peer),
		rpcPeers: make(map[string]*Peer),
		peersCheckQueue: list.New(),
		rpcCheckQueue: list.New(),

		bestRpcNodes: list.New(),
		bestRpcNodesLatency: 100,
	}
}

/*
func loadLocal() {
	if data, err := ioutil.ReadFile("near-drpc.json"); err == nil {
		_ = json.Unmarshal(data, &rpcPeers)
	}
}

func saveLocal() {
	defer func() {

	}()
	if data, err := json.Marshal(rpcPeers);  err == nil {
		_ = ioutil.WriteFile("near-drpc.json", data, 0666)
	}
}
*/

func SortPeers(peers map[string]*Peer) Peers {
	tmpMap := make(map[string]*Peer)
	for key, peer := range peers {
		tmpMap[key] = peer
	}
	peersSorted := make(Peers, len(tmpMap))
	idx := 0
	for _, peer := range tmpMap {
		peersSorted[idx] = peer
		idx++
	}
	sort.Sort(peersSorted)
	return peersSorted
}

func (near *Near)UpdateBestNodeByLatency() {
	near.locker.Lock()
	defer near.locker.Unlock()
	if len(near.rpcPeers) == 0 {
		return
	}

	peers := SortPeers(near.rpcPeers)

	near.bestRpcNodes = list.New()
	maxLen := len(peers)
	if maxLen > 0 {
		Latency := peers[0].Latency
		for i := 0; i < maxLen; i++ {
			if 	peers[i].Latency - Latency <= near.bestRpcNodesLatency {
				near.bestRpcNodes.PushBack(peers[i])
			}
		}
	}
}


func (near *Near)SetBestNodeToProxyServerBackends() {
	if near.bestRpcNodes != nil {
		backends := make(map[string]int64)
		for e := near.bestRpcNodes.Front(); e != nil; e = e.Next() {
			peer := e.Value.(*Peer)
			if !peer.Syncing {
				backends[peer.RpcAddr] = 100
			}
		}
		if len(backends) > 0{
			logger.Info("Setting proxy server backends: ", len(backends))
			near.proxyServer.SetBackends(backends)
		}
	}
}



func (near *Near)Summary() map[string]interface{} {
	near.locker.Lock()
	defer near.locker.Unlock()

	result := make(map[string]interface{})
	result["total_peers"] = len(near.peers)
	result["total_rpc_nodes"] = len(near.rpcPeers)
	result["all_rpc_nodes"] = SortPeers(near.rpcPeers)

	result["total_rpc_backends"] = near.bestRpcNodes.Len()
	peersBackends := make(Peers, near.bestRpcNodes.Len())
	idx := 0
	for e := near.bestRpcNodes.Front(); e != nil; e = e.Next() {
		peer := e.Value.(*Peer)
		peersBackends[idx] = peer
		idx++
	}
	result["backend_rpc_nodes"] = peersBackends
	return result
}
