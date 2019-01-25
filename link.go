package wikicrawl

type Link = string

// Unique set of url links.
type LinkSet struct {
	Set map[Link]bool
}

func (ls LinkSet) Add(link Link) bool {
	_, found := ls.Set[link]
	ls.Set[link] = true
	return !found
}

func NewLinkSet() LinkSet {
	return LinkSet{make(map[Link]bool, 1)}
}
