package datas

import (
	"github.com/attic-labs/noms/chunks"
	. "github.com/attic-labs/noms/dbg"
	"github.com/attic-labs/noms/ref"
	"github.com/attic-labs/noms/types"
)

//go:generate go run gen/types.go -o types.go

type DataStore struct {
	chunks.ChunkStore

	rt    chunks.RootTracker
	rc    *rootCache
	roots RootSet
}

func NewDataStore(cs chunks.ChunkStore, rt chunks.RootTracker) DataStore {
	return newDataStoreInternal(cs, rt, NewRootCache(cs))
}

func newDataStoreInternal(cs chunks.ChunkStore, rt chunks.RootTracker, rc *rootCache) DataStore {
	var roots RootSet
	rootRef := rt.Root()
	if (rootRef == ref.Ref{}) {
		roots = NewRootSet()
	} else {
		// BUG 11: This reads the entire database into memory. Whoopsie.
		roots = RootSetFromVal(types.MustReadValue(rootRef, cs).(types.Set))
	}
	return DataStore{
		cs, rt, rc, roots,
	}
}

func (ds *DataStore) Roots() RootSet {
	return ds.roots
}

func (ds *DataStore) Commit(newRoots RootSet) DataStore {
	Chk.True(newRoots.Len() > 0)

	parentsList := make([]types.Set, newRoots.Len())
	i := uint64(0)
	newRoots.Iter(func(root Root) (stop bool) {
		parentsList[i] = root.Parents()
		i++
		return
	})

	superceded := types.NewSet().Union(parentsList...)
	for !ds.doCommit(newRoots, RootSet{superceded}) {
	}

	return newDataStoreInternal(ds.ChunkStore, ds.rt, ds.rc)
}

func (ds *DataStore) doCommit(add, remove RootSet) bool {
	oldRootRef := ds.rt.Root()
	oldRoots := ds.Roots()

	prexisting := make([]Root, 0)
	ds.rc.Update(oldRootRef)
	add.Iter(func(r Root) (stop bool) {
		if ds.rc.Contains(r.Ref()) {
			prexisting = append(prexisting, r)
		}
		return
	})
	add = add.Remove(prexisting...)
	if add.Len() == 0 {
		return true
	}

	newRoots := oldRoots.Subtract(remove).Union(add)

	// TODO: This set will be orphaned if this UpdateRoot belnow fails
	newRootRef, err := types.WriteValue(newRoots.NomsValue(), ds)
	Chk.NoError(err)

	return ds.rt.UpdateRoot(newRootRef, oldRootRef)
}
