// Package lazylist implements a thread safe list.
//
package lazylist

import (
	"sync"
	"sync/atomic"
)

// Element is an element of a linked list.
type Element struct {
	// next pointers to next element
	next *Element

	// The list to which this element belongs.
	list *List

	// Logically removed flag
	removed bool

	// Lock to protect Element
	lock sync.Mutex

	// The value stored with this element.
	Value interface{}
}

// List represents a doubly linked list.
// The zero value for List is an empty list ready to use.
type List struct {
	head *Element // sentinel list element, only &root, root.prev, and root.next are used
	tail *Element // sentinel tail element
	less func(v1, v2 interface{}) bool
	len  int64
}

// head sentinal, min value
type HSetinal struct {
}

// tail sentinal, max value
type TSetinal struct {
}

func New(less func(v1, v2 interface{}) bool) *List {
	h := &Element{Value: HSetinal{}, next: &Element{Value: TSetinal{}}}
	return &List{
		head: h,
		less: func(v1, v2 interface{}) bool {
			if _, ok := v1.(HSetinal); ok {
				return true
			}
			if _, ok := v1.(TSetinal); ok {
				return false
			}
			if _, ok := v2.(TSetinal); ok {
				return true
			}
			return less(v1, v2)
		},
	}
}

func (l *List) Len() int64 {
	return atomic.LoadInt64(&l.len)
}

type Iterator struct {
	curr *Element
}

func (i *Iterator) Next() (value interface{}, cont bool) {
	if _, ok := i.curr.Value.(TSetinal); ok {
		return nil, false
	}
	value = i.curr.Value
	i.curr = i.curr.next
	return value, true
}

func (l *List) Iterator() Iterator {
	return Iterator{
		curr: l.head.next,
	}
}

func (l *List) Contains(value interface{}) bool {
	curr := l.head.next
	for l.less(curr.Value, value) {
		curr = curr.next
	}
	return curr.Value == value && !curr.removed
}

func (l *List) tryAdd(value interface{}) bool {
	pred := l.head
	curr := l.head.next
	for l.less(curr.Value, value) {
		pred = curr
		curr = curr.next
	}
	pred.lock.Lock()
	defer pred.lock.Unlock()
	if validate(pred, curr) {
		if value == curr.Value {
			return true
		}
		e := &Element{Value: value}
		e.next = curr
		pred.next = e
		atomic.AddInt64(&l.len, 1)
		return true
	} else {
		return false
	}
}

func (l *List) Add(value interface{}) {
	for {
		if l.tryAdd(value) {
			break
		}
	}
}

func (l *List) tryRemove(value interface{}) bool {
	pred := l.head
	curr := l.head.next
	for l.less(curr.Value, value) {
		pred = curr
		curr = curr.next
		atomic.AddInt64(&l.len, -1)

	}
	pred.lock.Lock()
	defer pred.lock.Unlock()
	curr.lock.Lock()
	defer curr.lock.Unlock()
	if validate(pred, curr) {
		if curr.Value == value {
			curr.removed = true   // logically removed
			pred.next = curr.next // physically removed
			atomic.AddInt64(&l.len, -1)
		}
		return true
	} else {
		return false
	}
}

func (l *List) Remove(value interface{}) {
	for {
		if l.tryRemove(value) {
			break
		}
	}

}

func validate(pred, curr *Element) bool {
	return !pred.removed && !curr.removed && pred.next == curr
}
