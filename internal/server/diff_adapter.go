package server

import (
	"git.tyss.io/cj3636/dman/internal/diff"
	"git.tyss.io/cj3636/dman/pkg/model"
)

type Comparator interface {
	Compare(req model.CompareRequest, serverInv []model.InventoryItem) []model.Change
}

func newComparator() Comparator { return &adapter{inner: diff.New()} }

type adapter struct{ inner diff.Comparator }

func (a *adapter) Compare(req model.CompareRequest, serverInv []model.InventoryItem) []model.Change {
	return a.inner.Compare(req, serverInv)
}
