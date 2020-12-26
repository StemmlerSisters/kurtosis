/*
 * Copyright (c) 2020 - present Kurtosis Technologies LLC.
 * All Rights Reserved.
 */

package partitioning

import (
	"github.com/kurtosis-tech/kurtosis/api_container/api"
)

// Stupid Go... why can't it just have generics so all this is part of stdlib
type ServiceIDSet struct {
	elems map[api.ServiceID]bool
}

func NewServiceIDSet() *ServiceIDSet {
	return &ServiceIDSet{
		elems: map[api.ServiceID]bool{},
	}
}

func (set *ServiceIDSet) AddElem(elem api.ServiceID) {
	set.elems[elem] = true
}

func (set *ServiceIDSet) AddElems(elems *ServiceIDSet) {
	for _, elem := range elems.Elems() {
		set.AddElem(elem)
	}
}

func (set ServiceIDSet) Copy() *ServiceIDSet {
	elemsCopy := map[api.ServiceID]bool{}
	for elem, _ := range set.elems {
		elemsCopy[elem] = true
	}
	return &ServiceIDSet{elems: elemsCopy}
}

func (set ServiceIDSet) Equals(other *ServiceIDSet) bool {
	if set.Size() != other.Size() {
		return false
	}
	for elem, _ := range set.elems {
		if !other.Contains(elem) {
			return false
		}
	}
	return true

}

func (set ServiceIDSet) Size() int {
	return len(set.elems)
}

func (set ServiceIDSet) Contains(elem api.ServiceID) bool {
	_, found := set.elems[elem]
	return found
}

func (set ServiceIDSet) Elems() []api.ServiceID {
	result := []api.ServiceID{}
	for elem, _ := range set.elems {
		result = append(result, elem)
	}
	return result
}

func (set *ServiceIDSet) RemoveElem(elem api.ServiceID) {
	delete(set.elems, elem)
}

// Removes the given elems from the set, if they exist
func (set *ServiceIDSet) RemoveElems(elems *ServiceIDSet) {
	for _, elem := range elems.Elems() {
		set.RemoveElem(elem)
	}
}