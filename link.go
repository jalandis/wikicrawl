package wikicrawl

import (
	"sync"
)

type Link = string

// Unique set of url links.
type LinkSet struct {
	sync.RWMutex

	Set map[Link]bool
}

func (ls *LinkSet) Add(link Link) bool {
	ls.Lock()
	defer ls.Unlock()

	_, found := ls.Set[link]
	if !found {
		ls.Set[link] = true
	}

	return !found
}

func (ls *LinkSet) Contains(link Link) bool {
	ls.RLock()
	defer ls.RUnlock()
	return ls.Set[link]
}

func NewLinkSet() LinkSet {
	return LinkSet{Set: make(map[Link]bool, 1)}
}
