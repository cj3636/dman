package diff

import "git.tyss.io/cj3636/dman/pkg/model"

type Comparator interface {
	Compare(req model.CompareRequest, serverInv []model.InventoryItem) []model.Change
}

type comparator struct{}

func New() Comparator { return &comparator{} }

func (c *comparator) Compare(req model.CompareRequest, serverInv []model.InventoryItem) []model.Change {
	clientMap := map[string]model.InventoryItem{}
	for _, it := range req.Inventory {
		clientMap[key(it)] = it
	}
	serverMap := map[string]model.InventoryItem{}
	for _, it := range serverInv {
		serverMap[key(it)] = it
	}
	var changes []model.Change
	// detect adds/modifies (client side)
	for k, cit := range clientMap {
		if sit, ok := serverMap[k]; !ok {
			changes = append(changes, model.Change{User: cit.User, Path: cit.Path, Type: model.ChangeAdd})
		} else if sit.Hash != cit.Hash {
			changes = append(changes, model.Change{User: cit.User, Path: cit.Path, Type: model.ChangeModify})
		}
	}
	// detect deletes (server has file missing locally)
	for k, sit := range serverMap {
		if _, ok := clientMap[k]; !ok {
			changes = append(changes, model.Change{User: sit.User, Path: sit.Path, Type: model.ChangeDelete})
		}
	}
	return changes
}

func key(it model.InventoryItem) string { return it.User + "::" + it.Path }
