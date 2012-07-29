package chunk

import (
	"github.com/Nightgunner5/stuzzd/block"
	"sort"
)

type SectionList []*Section

var _ sort.Interface = (*SectionList)(nil)

func (s *SectionList) Len() int {
	return len(*s)
}
func (s *SectionList) Less(a, b int) bool {
	return (*s)[a].Y < (*s)[b].Y
}
func (s *SectionList) Swap(a, b int) {
	(*s)[a], (*s)[b] = (*s)[b], (*s)[a]
}

func (s *SectionList) index(y byte) int {
	return sort.Search(len(*s), func(i int) bool {
		return (*s)[i].Y >= y
	})
}

// Determines if a section exists in O(log n) time.
func (s *SectionList) Has(y byte) bool {
	i := s.index(y)
	return i < len(*s) && (*s)[i].Y == y
}

// Get the section with a given Y, creating it if it is not already in the list.
// The caller must handle synchronization.
func (s *SectionList) Get(y byte) *Section {
	i := s.index(y)
	if i < len(*s) && (*s)[i].Y == y {
		return (*s)[i]
	}

	sec := &Section{Y: y}
	*s = append(*s, sec)
	sort.Sort(s)
	return sec
}

// Removes the section at the given Y coordinate if it has no non-air blocks.
func (s *SectionList) Compact(y byte) {
	i := s.index(y)
	if i >= len(*s) || (*s)[i].Y != y {
		return
	}

	for _, b := range (*s)[i].Blocks {
		if b != block.Air {
			return
		}
	}

	if i == 0 {
		*s = (*s)[1:]
	} else if i == len(*s)-1 {
		*s = (*s)[:len(*s)-1]
	} else {
		*s = append((*s)[:i-1], (*s)[i+1:]...)
	}
}

type Section struct {
	Y byte

	Blocks block.BlockSection
	Data   block.NibbleSection

	SkyLight   block.NibbleSection
	BlockLight block.NibbleSection
}
