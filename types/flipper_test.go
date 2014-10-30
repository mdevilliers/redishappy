package types

import (
	"sort"
	"testing"
)

func TestBasicMasterDetailsCollectionOperations(t *testing.T) {

	collection := NewMasterDetailsCollection()
	detail := &MasterDetails{ExternalPort: 1111, Name: "one", Ip: "1.1.1.1.", Port: 2222}
	collection.AddOrReplace(detail)

	items := collection.Items()

	if len(items) != 1 {
		t.Error("Wrong number of itmes in collection")
	}

	if items[0] != detail {
		t.Error("Should return reference to details")
	}
}

func TestReplaceMasterDetailsCollection(t *testing.T) {

	collection := NewMasterDetailsCollection()
	detail1 := &MasterDetails{ExternalPort: 1111, Name: "one", Ip: "1.1.1.1.", Port: 2222}
	detail2 := &MasterDetails{ExternalPort: 2222, Name: "one", Ip: "2.2.2.2.", Port: 3333}
	collection.AddOrReplace(detail1)
	collection.AddOrReplace(detail2)

	items := collection.Items()

	if len(items) != 1 {
		t.Error("Wrong number of itmes in collection")
	}

	if items[0] != detail2 {
		t.Error("Should return reference to details2")
	}
}

func TestSortMasterDetailsByName(t *testing.T) {

	collection := NewMasterDetailsCollection()

	detail1 := &MasterDetails{ExternalPort: 1111, Name: "a", Ip: "1.1.1.1.", Port: 2222}
	detail2 := &MasterDetails{ExternalPort: 2222, Name: "b", Ip: "2.2.2.2.", Port: 3333}
	detail3 := &MasterDetails{ExternalPort: 2222, Name: "c", Ip: "2.2.2.2.", Port: 3333}

	collection.AddOrReplace(detail3)
	collection.AddOrReplace(detail2)
	collection.AddOrReplace(detail1)

	items := collection.Items()
	sort.Sort(ByName(items))

	if len(items) != 3 {
		t.Error("Wrong number of itmes in collection")
	}

	if items[0] != detail1 {
		t.Error("Should return reference to details2")
	}
}

func TestMasterDetailsIsEmpty(t *testing.T) {

	collection := NewMasterDetailsCollection()

	if collection.IsEmpty() == false {
		t.Error("Collection should be empty")
	}

	detail := &MasterDetails{ExternalPort: 1111, Name: "a", Ip: "1.1.1.1.", Port: 2222}
	collection.AddOrReplace(detail)

	if collection.IsEmpty() {
		t.Error("Collection should not be empty")
	}
}
