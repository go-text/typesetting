package fontscan

import (
	"container/list"
)

// this file implements the family substitution feature,
// inspired by fontconfig.
// it works by defining a set of modifications to apply
// to a user provided family
// each of them may happen one (or more) alternative family to
// look for

// we want to easily insert at the start,
// the end and "around" an element
type familyList struct {
	*list.List
}

// panic if `families` is empty
func newFamilyList(families []string) familyList {
	var out list.List
	for _, s := range families {
		out.PushBack(s)
	}
	return familyList{List: &out}
}

// returns the node for `family` or nil, if not found
func (fl familyList) contains(family string) *list.Element {
	for l := fl.List.Front(); l != nil; l = l.Next() {
		if l.Value.(string) == family {
			return l
		}
	}
	return nil
}

// return the crible corresponding to the order
func (fl familyList) compile() familyCrible {
	out := make(familyCrible)
	i := 0
	for l := fl.List.Front(); l != nil; l, i = l.Next(), i+1 {
		out[l.Value.(string)] = i
	}
	return out
}

func (fl familyList) insertStart(families []string) {
	L := len(families)
	for i := range families {
		fl.List.PushFront(families[L-1-i])
	}
}

func (fl familyList) insertEnd(families []string) {
	for _, s := range families {
		fl.List.PushBack(s)
	}
}

// insertAfter inserts families right after element
func (fl familyList) insertAfter(element *list.Element, families []string) {
	for _, s := range families {
		element = fl.List.InsertAfter(s, element)
	}
}

// insertBefore inserts families right before element
func (fl familyList) insertBefore(element *list.Element, families []string) {
	L := len(families)
	for i := range families {
		element = fl.List.InsertBefore(families[L-1-i], element)
	}
}

func (fl familyList) replace(element *list.Element, families []string) {
	fl.insertAfter(element, families)
	fl.List.Remove(element)
}

// ----- substitutions ------

type substitutionOp uint8

const (
	opAppend substitutionOp = iota
	opAppendLast
	opPrepend
	opPrependFirst
	opReplace
)

type substitution struct {
	targetFamily       string         // the family concerned
	additionalFamilies []string       // the families to add
	op                 substitutionOp // how to insert the families
}

func (fl familyList) execute(subs substitution) {
	element := fl.contains(subs.targetFamily)
	if element == nil {
		return
	}

	switch subs.op {
	case opAppend:
		fl.insertAfter(element, subs.additionalFamilies)
	case opAppendLast:
		fl.insertEnd(subs.additionalFamilies)
	case opPrepend:
		fl.insertBefore(element, subs.additionalFamilies)
	case opPrependFirst:
		fl.insertStart(subs.additionalFamilies)
	case opReplace:
		fl.replace(element, subs.additionalFamilies)
	default:
		panic("exhaustive switch")
	}
}
