package matechers

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestJQMatcher(t *testing.T) {
	g := NewWithT(t)

	g.Expect(`{"a":1}`).Should(MatchJQ(`.a == 1`))
	g.Expect(`{"a":1}`).Should(Not(MatchJQ(`.a == 2`)))
}
