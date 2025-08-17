// Inspired by github.com/SaveTheRbtz/generic-sync-map-go but technically
// written from scratch with Go 1.23's sync.Map.
// Copyright 2024 Runxi Yu (porting it to generics)
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmap

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

// ComparableMap[K comparable, V comparable] is like a Go map[K]V but is safe for concurrent use
// by multiple goroutines without additional locking or coordination.  Loads,
// stores, and deletes run in amortized constant time.
//
// The ComparableMap type is optimized for two common use cases: (1) when the comparableEntry for a given
// key is only ever written once but read many times, as in caches that only grow,
// or (2) when multiple goroutines read, write, and overwrite entries for disjoint
// sets of keys. In these two cases, use of a ComparableMap may significantly reduce lock
// contention compared to a Go map paired with a separate [Mutex] or [RWMutex].
//
// The zero ComparableMap is empty and ready for use. A ComparableMap must not be copied after first use.
//
// In the terminology of [the Go memory model], ComparableMap arranges that a write operation
// “synchronizes before” any read operation that observes the effect of the write, where
// read and write operations are defined as follows.
// [ComparableMap.Load], [ComparableMap.LoadAndDelete], [ComparableMap.LoadOrStore], [ComparableMap.Swap], [ComparableMap.CompareAndSwap],
// and [ComparableMap.CompareAndDelete] are read operations;
// [ComparableMap.Delete], [ComparableMap.LoadAndDelete], [ComparableMap.Store], and [ComparableMap.Swap] are write operations;
// [ComparableMap.LoadOrStore] is a write operation when it returns loaded set to false;
// [ComparableMap.CompareAndSwap] is a write operation when it returns swapped set to true;
// and [ComparableMap.CompareAndDelete] is a write operation when it returns deleted set to true.
//
// [the Go memory model]: https://go.dev/ref/mem
type ComparableMap[K comparable, V comparable] struct {
	mu sync.Mutex

	// read contains the portion of the map's contents that are safe for
	// concurrent access (with or without mu held).
	//
	// The read field itself is always safe to load, but must only be stored with
	// mu held.
	//
	// Entries stored in read may be updated concurrently without mu, but updating
	// a previously-comparableExpunged comparableEntry requires that the comparableEntry be copied to the dirty
	// map and uncomparableExpunged with mu held.
	read atomic.Pointer[comparableReadOnly[K, V]]

	// dirty contains the portion of the map's contents that require mu to be
	// held. To ensure that the dirty map can be promoted to the read map quickly,
	// it also includes all of the non-comparableExpunged entries in the read map.
	//
	// Expunged entries are not stored in the dirty map. An comparableExpunged comparableEntry in the
	// clean map must be uncomparableExpunged and added to the dirty map before a new value
	// can be stored to it.
	//
	// If the dirty map is nil, the next write to the map will initialize it by
	// making a shallow copy of the clean map, omitting stale entries.
	dirty map[K]*comparableEntry[V]

	// misses counts the number of loads since the read map was last updated that
	// needed to lock mu to determine whether the key was present.
	//
	// Once enough misses have occurred to cover the cost of copying the dirty
	// map, the dirty map will be promoted to the read map (in the unamended
	// state) and the next store to the map will make a new dirty copy.
	misses int
}

// comparableReadOnly is an immutable struct stored atomically in the ComparableMap.read field.
type comparableReadOnly[K comparable, V comparable] struct {
	m       map[K]*comparableEntry[V]
	amended bool // true if the dirty map contains some key not in m.
}

// comparableExpunged is an arbitrary pointer that marks entries which have been deleted
// from the dirty map.
var comparableExpunged = unsafe.Pointer(new(any))

// An comparableEntry is a slot in the map corresponding to a particular key.
type comparableEntry[V comparable] struct {
	// p points to the value stored for the comparableEntry.
	//
	// If p == nil, the comparableEntry has been deleted, and either m.dirty == nil or
	// m.dirty[key] is e.
	//
	// If p == comparableExpunged, the comparableEntry has been deleted, m.dirty != nil, and the comparableEntry
	// is missing from m.dirty.
	//
	// Otherwise, the comparableEntry is valid and recorded in m.read.m[key] and, if m.dirty
	// != nil, in m.dirty[key].
	//
	// An comparableEntry can be deleted by atomic replacement with nil: when m.dirty is
	// next created, it will atomically replace nil with comparableExpunged and leave
	// m.dirty[key] unset.
	//
	// An comparableEntry's associated value can be updated by atomic replacement, provided
	// p != comparableExpunged. If p == comparableExpunged, an comparableEntry's associated value can be updated
	// only after first setting m.dirty[key] = e so that lookups using the dirty
	// map find the comparableEntry.
	p unsafe.Pointer
}

func newComparableEntry[V comparable](i V) *comparableEntry[V] {
	return &comparableEntry[V]{p: unsafe.Pointer(&i)}
}

func (m *ComparableMap[K, V]) loadReadOnly() comparableReadOnly[K, V] {
	if p := m.read.Load(); p != nil {
		return *p
	}
	return comparableReadOnly[K, V]{}
}

// Load returns the value stored in the map for a key, or nil if no
// value is present.
// The ok result indicates whether value was found in the map.
func (m *ComparableMap[K, V]) Load(key K) (value V, ok bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		// Avoid reporting a spurious miss if m.dirty got promoted while we were
		// blocked on m.mu. (If further loads of the same key will not miss, it's
		// not worth copying the dirty map for this key.)
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			// Regardless of whether the comparableEntry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		return *new(V), false
	}
	return e.load()
}

func (e *comparableEntry[V]) load() (value V, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == comparableExpunged {
		return value, false
	}
	return *(*V)(p), true
}

// Store sets the value for a key.
func (m *ComparableMap[K, V]) Store(key K, value V) {
	_, _ = m.Swap(key, value)
}

// Clear deletes all the entries, resulting in an empty ComparableMap.
func (m *ComparableMap[K, V]) Clear() {
	read := m.loadReadOnly()
	if len(read.m) == 0 && !read.amended {
		// Avoid allocating a new comparableReadOnly when the map is already clear.
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	read = m.loadReadOnly()
	if len(read.m) > 0 || read.amended {
		m.read.Store(&comparableReadOnly[K, V]{})
	}

	clear(m.dirty)
	// Don't immediately promote the newly-cleared dirty map on the next operation.
	m.misses = 0
}

// tryCompareAndSwap compare the comparableEntry with the given old value and swaps
// it with a new value if the comparableEntry is equal to the old value, and the comparableEntry
// has not been comparableExpunged.
//
// If the comparableEntry is comparableExpunged, tryCompareAndSwap returns false and leaves
// the comparableEntry unchanged.
func (e *comparableEntry[V]) tryCompareAndSwap(old V, new V) bool {
	p := atomic.LoadPointer(&e.p)
	if p == nil || p == comparableExpunged || *(*V)(p) != old { // XXX
		return false
	}

	// Copy the pointer after the first load to make this method more amenable
	// to escape analysis: if the comparison fails from the start, we shouldn't
	// bother heap-allocating a pointer to store.
	nc := new
	for {
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(&nc)) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
		if p == nil || p == comparableExpunged || *(*V)(p) != old {
			return false
		}
	}
}

// unexpungeLocked ensures that the comparableEntry is not marked as comparableExpunged.
//
// If the comparableEntry was previously comparableExpunged, it must be added to the dirty map
// before m.mu is unlocked.
func (e *comparableEntry[V]) unexpungeLocked() (wasExpunged bool) {
	return atomic.CompareAndSwapPointer(&e.p, comparableExpunged, nil)
}

// swapLocked unconditionally swaps a value into the comparableEntry.
//
// The comparableEntry must be known not to be comparableExpunged.
func (e *comparableEntry[V]) swapLocked(i *V) *V {
	return (*V)(atomic.SwapPointer(&e.p, unsafe.Pointer(i)))
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *ComparableMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	// Avoid locking if it's a clean hit.
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	m.mu.Lock()
	read = m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			m.dirty[key] = e
		}
		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := m.dirty[key]; ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		m.missLocked()
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(&comparableReadOnly[K, V]{m: read.m, amended: true})
		}
		m.dirty[key] = newComparableEntry(value)
		actual, loaded = value, false
	}
	m.mu.Unlock()

	return actual, loaded
}

// tryLoadOrStore atomically loads or stores a value if the comparableEntry is not
// comparableExpunged.
//
// If the comparableEntry is comparableExpunged, tryLoadOrStore leaves the comparableEntry unchanged and
// returns with ok==false.
func (e *comparableEntry[V]) tryLoadOrStore(i V) (actual V, loaded, ok bool) {
	p := atomic.LoadPointer(&e.p)
	if p == comparableExpunged {
		return actual, false, false
	}
	if p != nil {
		return *(*V)(p), true, true
	}

	// Copy the pointer after the first load to make this method more amenable
	// to escape analysis: if we hit the "load" path or the comparableEntry is comparableExpunged, we
	// shouldn't bother heap-allocating.
	ic := i
	for {
		if atomic.CompareAndSwapPointer(&e.p, nil, unsafe.Pointer(&ic)) {
			return i, false, true
		}
		p = atomic.LoadPointer(&e.p)
		if p == comparableExpunged {
			return actual, false, false
		}
		if p != nil {
			return *(*V)(p), true, true
		}
	}
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *ComparableMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			delete(m.dirty, key)
			// Regardless of whether the comparableEntry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if ok {
		return e.delete()
	}
	return value, false
}

// Delete deletes the value for a key.
func (m *ComparableMap[K, V]) Delete(key K) {
	m.LoadAndDelete(key)
}

func (e *comparableEntry[V]) delete() (value V, ok bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == comparableExpunged {
			return value, false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return *(*V)(p), true
		}
	}
}

// trySwap swaps a value if the comparableEntry has not been comparableExpunged.
//
// If the comparableEntry is comparableExpunged, trySwap returns false and leaves the comparableEntry
// unchanged.
func (e *comparableEntry[V]) trySwap(i *V) (*V, bool) {
	for {
		p := atomic.LoadPointer(&e.p)
		if p == comparableExpunged {
			return nil, false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, unsafe.Pointer(i)) {
			return (*V)(p), true
		}
	}
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *ComparableMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if v, ok := e.trySwap(&value); ok {
			if v == nil {
				return previous, false
			}
			return *v, true
		}
	}

	m.mu.Lock()
	read = m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			// The comparableEntry was previously comparableExpunged, which implies that there is a
			// non-nil dirty map and this comparableEntry is not in it.
			m.dirty[key] = e
		}
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else if e, ok := m.dirty[key]; ok {
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else {
		if !read.amended {
			// We're adding the first new key to the dirty map.
			// Make sure it is allocated and mark the read-only map as incomplete.
			m.dirtyLocked()
			m.read.Store(&comparableReadOnly[K, V]{m: read.m, amended: true})
		}
		m.dirty[key] = newComparableEntry(value)
	}
	m.mu.Unlock()
	return previous, loaded
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
func (m *ComparableMap[K, V]) CompareAndSwap(key K, old, new V) (swapped bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		return e.tryCompareAndSwap(old, new)
	} else if !read.amended {
		return false // No existing value for key.
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	read = m.loadReadOnly()
	swapped = false
	if e, ok := read.m[key]; ok {
		swapped = e.tryCompareAndSwap(old, new)
	} else if e, ok := m.dirty[key]; ok {
		swapped = e.tryCompareAndSwap(old, new)
		// We needed to lock mu in order to load the comparableEntry for key,
		// and the operation didn't change the set of keys in the map
		// (so it would be made more efficient by promoting the dirty
		// map to read-only).
		// Count it as a miss so that we will eventually switch to the
		// more efficient steady state.
		m.missLocked()
	}
	return swapped
}

// CompareAndDelete deletes the comparableEntry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// If there is no current value for key in the map, CompareAndDelete
// returns false (even if the old value is a nil pointer).
func (m *ComparableMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			// Don't delete key from m.dirty: we still need to do the “compare” part
			// of the operation. The comparableEntry will eventually be comparableExpunged when the
			// dirty map is promoted to the read map.
			//
			// Regardless of whether the comparableEntry was present, record a miss: this key
			// will take the slow path until the dirty map is promoted to the read
			// map.
			m.missLocked()
		}
		m.mu.Unlock()
	}
	for ok {
		p := atomic.LoadPointer(&e.p)
		if p == nil || p == comparableExpunged || *(*V)(p) != old {
			return false
		}
		if atomic.CompareAndSwapPointer(&e.p, p, nil) {
			return true
		}
	}
	return false
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
//
// Range does not necessarily correspond to any consistent snapshot of the ComparableMap's
// contents: no key will be visited more than once, but if the value for any key
// is stored or deleted concurrently (including by f), Range may reflect any
// mapping for that key from any point during the Range call. Range does not
// block other methods on the receiver; even f itself may call any method on m.
//
// Range may be O(N) with the number of elements in the map even if f returns
// false after a constant number of calls.
func (m *ComparableMap[K, V]) Range(f func(key K, value V) bool) {
	// We need to be able to iterate over all of the keys that were already
	// present at the start of the call to Range.
	// If read.amended is false, then read.m satisfies that property without
	// requiring us to hold m.mu for a long time.
	read := m.loadReadOnly()
	if read.amended {
		// m.dirty contains keys not in read.m. Fortunately, Range is already O(N)
		// (assuming the caller does not break out early), so a call to Range
		// amortizes an entire copy of the map: we can promote the dirty copy
		// immediately!
		m.mu.Lock()
		read = m.loadReadOnly()
		if read.amended {
			read = comparableReadOnly[K, V]{m: m.dirty}
			copyRead := read
			m.read.Store(&copyRead)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

func (m *ComparableMap[K, V]) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(&comparableReadOnly[K, V]{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

func (m *ComparableMap[K, V]) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read := m.loadReadOnly()
	m.dirty = make(map[K]*comparableEntry[V], len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

func (e *comparableEntry[V]) tryExpungeLocked() (isExpunged bool) {
	p := atomic.LoadPointer(&e.p)
	for p == nil {
		if atomic.CompareAndSwapPointer(&e.p, nil, comparableExpunged) {
			return true
		}
		p = atomic.LoadPointer(&e.p)
	}
	return p == comparableExpunged
}
