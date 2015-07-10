package dataset

import (
	"github.com/attic-labs/noms/datas"
	. "github.com/attic-labs/noms/dbg"
	"github.com/attic-labs/noms/types"
)

//go:generate go run gen/types.go -o types.go

func GetDatasets(ds datas.DataStore) DatasetSet {
	if ds.Roots().Empty() {
		return NewDatasetSet()
	} else {
		// BUG 13: We don't ever want to branch the datasets database. Currently we can't avoid that, but we should change DataStore::Commit() to support that mode of operation.
		Chk.EqualValues(1, ds.Roots().Len())
		return DatasetSetFromVal(ds.Roots().Any().Value())
	}
}

func CommitDatasets(ds datas.DataStore, datasets DatasetSet) datas.DataStore {
	return ds.Commit(datas.NewRootSet().Insert(
		datas.NewRoot().SetParents(
			ds.Roots().NomsValue()).SetValue(
			datasets.NomsValue())))
}

func getDataset(datasets DatasetSet, datasetID string) (r *Dataset) {
	datasets.Iter(func(dataset Dataset) (stop bool) {
		if dataset.Id().String() == datasetID {
			r = &dataset
			stop = true
		}
		return
	})
	return
}

func GetDatasetRoot(datasets DatasetSet, datasetID string) types.Value {
	dataset := getDataset(datasets, datasetID)
	if dataset == nil {
		return nil
	}
	return dataset.Root()
}

func SetDatasetRoot(datasets DatasetSet, datasetID string, val types.Value) DatasetSet {
	newDataset := NewDataset().SetId(types.NewString(datasetID)).SetRoot(val)
	dataset := getDataset(datasets, datasetID)
	if dataset == nil {
		return datasets.Insert(newDataset)
	}
	return datasets.Remove(*dataset).Insert(newDataset)
}
