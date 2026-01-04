package diff

import (
	"math/rand"
	"testing"
	"time"

	"git.tyss.io/cj3636/dman/pkg/model"
)

// property: no duplicate (user,path) pairs in output; modify implies different hash.
func TestComparatorProperties(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	c := New()
	for i := 0; i < 200; i++ {
		clientInv, serverInv := randomInventories(r)
		req := model.CompareRequest{Users: []string{"u"}, Inventory: clientInv}
		changes := c.Compare(req, serverInv)
		seen := map[string]struct{}{}
		for _, ch := range changes {
			k := ch.User + "::" + ch.Path
			if _, ok := seen[k]; ok {
				t.Fatalf("duplicate change for %s", k)
			}
			seen[k] = struct{}{}
			if ch.Type == model.ChangeModify {
				var clientHash, serverHash string
				for _, it := range clientInv {
					if it.User == ch.User && it.Path == ch.Path {
						clientHash = it.Hash
					}
				}
				for _, it := range serverInv {
					if it.User == ch.User && it.Path == ch.Path {
						serverHash = it.Hash
					}
				}
				if clientHash == serverHash {
					t.Fatalf("modify with identical hash %s", ch.Path)
				}
			}
		}
	}
}

func randomInventories(r *rand.Rand) ([]model.InventoryItem, []model.InventoryItem) {
	n := r.Intn(10)
	client := make([]model.InventoryItem, 0, n)
	server := make([]model.InventoryItem, 0, n)
	for i := 0; i < n; i++ {
		path := randomPath(r)
		chash := randHash(r)
		same := r.Intn(2) == 0
		shash := chash
		if !same {
			shash = randHash(r)
		}
		client = append(client, model.InventoryItem{User: "u", Path: path, Hash: chash})
		// randomly drop from server or include modified
		if r.Intn(3) != 0 {
			server = append(server, model.InventoryItem{User: "u", Path: path, Hash: shash})
		}
		// maybe extra server-only file
		if r.Intn(5) == 0 {
			server = append(server, model.InventoryItem{User: "u", Path: randomPath(r), Hash: randHash(r)})
		}
	}
	return client, server
}

func randHash(r *rand.Rand) string {
	letters := []rune("abcdef0123456789")
	b := make([]rune, 8)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func randomPath(r *rand.Rand) string {
	return "f" + randHash(r) + ".txt"
}
