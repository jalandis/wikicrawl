package wikicrawl

import (
	"testing"
)

func TestLinkSet(t *testing.T) {
	t.Run("Simple LinkSet data structure", func(t *testing.T) {
		t.Run("Enforces uniqueness", func(t *testing.T) {
			t.Parallel()

			found := NewLinkSet()
			found.Add("1")

			if found.Add("1") {
				t.Errorf("LinkSet Add should report false for duplicates.")
			}

			if len(found.Set) != 1 {
				t.Errorf("LinkSet Add should not include duplicates.")
			}
		})
	})
}
