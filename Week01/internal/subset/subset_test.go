package subset

import (
	mapset "github.com/deckarep/golang-set"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestSubset(t *testing.T) {
	inss0 := make([]int, 100)
	for i := 0; i < len(inss0); i++ {
		inss0[i] = i + 1
	}
	size := 5
	subset0 := Subset(inss0, size)
	sort.Ints(subset0)
	t.Logf("Origin subset of size %d: %v", size, subset0)

	set0 := mapset.NewSet()
	for _, i := range subset0 {
		set0.Add(i)
	}
	ra := rand.New(rand.NewSource(time.Now().UnixNano()))

	for times := 0; times < 10; times++ {
		// lost backends
		inss1 := make([]int, 0, len(inss0))
		inss1 = append(inss1, inss0...)
		losts := 2
		for i := 0; i < losts; i++ {
			index := ra.Intn(len(inss1))
			inss1[index] = inss1[len(inss1)-1]
			inss1 = inss1[:len(inss1)-1]
		}
		subset1 := Subset(inss1, size)
		sort.Ints(subset1)
		t.Logf("%dth time: After %d backends gone, subset of size %d: %v", times, losts, size, subset1)

		set1 := mapset.NewSet()
		for _, i := range subset1 {
			set1.Add(i)
		}
		difference := set1.Difference(set0)
		t.Logf("%dth time: subset change %d", times, difference.Cardinality()) // 几乎都变了。。。不是说变化不大吗？
	}
}
