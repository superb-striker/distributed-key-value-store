package hashring 

import ( 
	"errors" 
	"fmt" 
	"hash/crc32" 
	"sort" 
	"sync" 
)

var ErrNoNodes = errors.New("hashring: no nodes in ring") 

const DefaultVirtualNodes = 150 

type Ring struct { 
	mu sync.RWMutex 
	virtualNodes int 
	sortedHashes []uint32 
	hashToNode map[uint32]string 
	nodes map[string]bool 
}

func New(vnodes int) *Ring { 
	if vnodes <= 0 { 
		vnodes = DefaultVirtualNodes 
	} 
	return &Ring{ 
		virtualNodes: vnodes, 
		hashToNode: make(map[uint32]string), 
		nodes: make(map[string]bool), 
	} 
} 

func hashKey(key string) uint32 { 
	return crc32.ChecksumIEEE([]byte(key)) 
} 

func vnodeKey(nodeID string, i int) string { 
	return fmt.Sprintf("%s#%d", nodeID, i) 
}

func (r *Ring) AddNode(nodeID string) { 
	r.mu.Lock() 
	defer r.mu.Unlock() 
	if r.nodes[nodeID] { 
		return 
	} 
	r.nodes[nodeID] = true 
	for i := 0; i < r.virtualNodes; i++ { 
		h := hashKey(vnodeKey(nodeID, i)) 
		r.hashToNode[h] = nodeID 
	} 
	r.rebuildSorted() 
}

func (r *Ring) RemoveNode(nodeID string) { 
	r.mu.Lock() 
	defer r.mu.Unlock() 
	if !r.nodes[nodeID] { 
		return 
	} 
	delete(r.nodes, nodeID) 
	for i := 0; i < r.virtualNodes; i++ { 
		h := hashKey(vnodeKey(nodeID, i)) 
		if r.hashToNode[h] == nodeID { 
			delete(r.hashToNode, h) 
		} 
	} 
	r.rebuildSorted() 
} 

func (r *Ring) rebuildSorted() { 
	hashes := make([]uint32, 0, len(r.hashToNode)) 
	for h := range r.hashToNode { 
		hashes = append(hashes, h) 
	} 
	sort.Slice(hashes, func(i, j int) bool { return hashes[i] < hashes[j] }) 
	r.sortedHashes = hashes 
} 

func (r *Ring) Owner(key string) (string, error) { 
	r.mu.RLock() 
	defer r.mu.RUnlock() 
	if len(r.sortedHashes) == 0 { 
		return "", ErrNoNodes 
	} 
	h := hashKey(key) 
	idx := sort.Search(len(r.sortedHashes), func(i int) bool { return r.sortedHashes[i] >= h }) 
	if idx == len(r.sortedHashes) {
		idx = 0 // wrap around the ring 
	} 
	return r.hashToNode[r.sortedHashes[idx]], nil 
}

