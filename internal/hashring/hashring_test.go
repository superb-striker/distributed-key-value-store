package hashring 

import ( 
	"fmt" 
	"testing" 
) 

func TestOwnerDeterministic(t *testing.T) { 
	r := New(DefaultVirtualNodes) 
	nodes := []string{"node-a", "node-b", "node-c"} 
	for _, n := range nodes { 
		r.AddNode(n) 
	} 
	nodeSet := map[string]bool{} 
	for _, n := range nodes { 
		nodeSet[n] = true 
	} 
	for i := 0; i < 1000; i++ { 
		key := fmt.Sprintf("key-%d", i) 
		owner, err := r.Owner(key) 
		if err != nil {
			t.Fatalf("Owner(%q) returned error: %v", key, err) 
		} 
		if !nodeSet[owner] { 
			t.Fatalf("Owner(%q) = %q, not a member of the ring", key, owner)   
		} 
		// same key, same ring -> should be same owner every time 
		owner2, _ := r.Owner(key) 
		if owner != owner2 { 
			t.Fatalf("Owner(%q) not deterministic: got %q then %q", key, owner, owner2) 
		} 
	} 
} 

func TestEmptyRing(t *testing.T) { 
	r := New(DefaultVirtualNodes) 
	if _, err := r.Owner("any-key"); err != ErrNoNodes { 
		t.Fatalf("expected ErrNoNodes on empty ring, got %v", err) 
	} 
} 

func TestDistributionIsReasonablyBalanced(t *testing.T) { 
	r := New(DefaultVirtualNodes) 
	nodes := []string{"node-a", "node-b", "node-c"} 
	for _, n := range nodes { 
		r.AddNode(n) 
	} 
	const numKeys = 30000 
	counts := map[string]int{} 
	for i := 0; i < numKeys; i++ { 
		owner, err := r.Owner(fmt.Sprintf("key-%d", i)) 
		if err != nil { 
			t.Fatalf("unexpected error: %v", err) 
		} 
	counts[owner]++ 
	} 

	expected := numKeys / len(nodes) 
	tolerance := float64(expected) * 0.20 
	for _, n := range nodes { 
		c := counts[n] 
		diff := float64(c - expected) 
		if diff < 0 { 
			diff = -diff 
		} 
		if diff > tolerance { 
			t.Errorf("node %s got %d keys, expected ~%d (+/- %.0f)", n, c, expected, tolerance) 
		} 
	} 
} 

func TestAddNodeRemapsApproximatelyOneOverN(t *testing.T) { 
	r := New(DefaultVirtualNodes) 
	initial := []string{"node-a", "node-b", "node-c"} 
	for _, n := range initial { 
		r.AddNode(n) 
	} 
	const numKeys = 30000 
	keys := make([]string, numKeys) 
	before := make([]string, numKeys) 
	for i := 0; i < numKeys; i++ { 
		keys[i] = fmt.Sprintf("key-%d", i) 
		owner, err := r.Owner(keys[i]) 
		if err != nil { 
			t.Fatalf("unexpected error: %v", err) 
		} 
		before[i] = owner 
	} 
	// Add a 4th node: N=3 -> N=4. Expected remap fraction ~= 1/4 = 25%. 
	r.AddNode("node-d") 
	moved := 0 
	for i, k := range keys { 
		owner, err := r.Owner(k) 
		if err != nil { 
			t.Fatalf("unexpected error: %v", err) 
		} 
		if owner != before[i] { 
			moved++ 
		} 
	} 
	fraction := float64(moved) / float64(numKeys) 
	if fraction < 0.15 || fraction > 0.35 { 
		t.Errorf("adding 1 node to a 3-node ring remapped %.1f%% of keys, want ~25%% (15%%-35%% band)", fraction*100) 
	} else { 
		t.Logf("adding 1 node to a 3-node ring remapped %.1f%% of keys (ideal 25%%)", fraction*100) 
	} 
} 

func TestRemoveNodeRemapsApproximatelyOneOverN(t *testing.T) { 
	r := New(DefaultVirtualNodes) 
	initial := []string{"node-a", "node-b", "node-c", "node-d"} 
	for _, n := range initial { 
		r.AddNode(n) 
	} 
	const numKeys = 30000 
	keys := make([]string, numKeys) 
	before := make([]string, numKeys) 
	for i := 0; i < numKeys; i++ { 
		keys[i] = fmt.Sprintf("key-%d", i) 
		owner, _ := r.Owner(keys[i]) 
		before[i] = owner 
	} 
	r.RemoveNode("node-d") 
	moved := 0 
	for i, k := range keys { 
		owner, err := r.Owner(k) 
		if err != nil { 
			t.Fatalf("unexpected error: %v", err) 
		} 
		if owner == "node-d" { 
			t.Fatalf("key %q still mapped to removed node-d", k) 
		} 
		if owner != before[i] { 
			moved++ 
		} 
	} 
	fraction := float64(moved) / float64(numKeys) 
	if fraction < 0.15 || fraction > 0.35 { 
		t.Errorf("removing 1 of 4 nodes remapped %.1f%% of keys, want ~25%% (15%%-35%% band)", fraction*100) 
	} else { 
		t.Logf("removing 1 of 4 nodes remapped %.1f%% of keys (ideal 25%%)", fraction*100) 
	} 
}
