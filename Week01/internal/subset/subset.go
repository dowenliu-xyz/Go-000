package subset

import (
	"github.com/dgryski/go-farm"
	"github.com/go-kratos/kratos/pkg/conf/env"
	"math/rand"
	"sort"
)

func Subset(inss []int, size int) []int {
	backends := inss
	if len(backends) <= size {
		return backends
	}
	clientID := env.Hostname
	sort.Ints(backends)
	count := len(backends) / size
	// hash得到id
	id := farm.Fingerprint64([]byte(clientID))
	// 获取 rand 轮数
	round := int64(id / uint64(count))

	s := rand.NewSource(round)
	ra := rand.New(s)
	// 根据source洗牌
	ra.Shuffle(len(backends), func(i, j int) {
		backends[i], backends[j] = backends[j], backends[i]
	})
	start := (id % uint64(count)) * uint64(size)
	return backends[int(start) : int(start)+size]
}
