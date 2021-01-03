package rolling

import (
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	nowFn = time.Now
)

// Bucket 滑动窗口中的各个桶。
type Bucket struct {
	// Success 桶时间片内成功请求计数
	Success int64
	// Failure 桶时间片内失败请求计数
	Failure int64
	// next 指向下一个桶的指针
	next *Bucket
}

// AddSuccess 原子性的增加成功计数
func (b *Bucket) AddSuccess(val int64) {
	atomic.AddInt64(&b.Success, val)
}

// AddFailure 原子性的增加失败计数
func (b *Bucket) AddFailure(val int64) {
	atomic.AddInt64(&b.Failure, val)
}

// Reset 重置桶
func (b *Bucket) Reset() {
	b.Success = 0
	b.Failure = 0
}

// Next 获取下一个桶
func (b *Bucket) Next() *Bucket {
	return b.next
}

// BucketRing 由桶构成的环结构，滑动窗口在其上单向滑动
type BucketRing struct {
	buckets []*Bucket
	current *Bucket
	size    int
}

func NewBucketRing(size int) *BucketRing {
	if size <= 0 {
		panic("bucket ring size must be greater than 0")
	}
	r := &BucketRing{
		buckets: make([]*Bucket, size),
		size:    size,
	}
	for i := 0; i < size; i++ {
		r.buckets[i] = &Bucket{}
	}
	r.current = r.buckets[0]
	for i := 0; i < size; i++ {
		r.buckets[i].next = r.buckets[(i+1)%size]
	}
	return r
}

// Rotate 向右轮换当前桶节点
func (r *BucketRing) Rotate() {
	r.current.next.Reset()
	r.current = r.current.next
}

// Current 获取当前焦点桶
func (r *BucketRing) Current() *Bucket {
	return r.current
}

// Window 滑动窗口。在 bucketRing 上滑动
type Window struct {
	ring       *BucketRing
	size       int
	duration   time.Duration
	internal   time.Duration
	left       *Bucket
	bucketTime time.Time
	mu         sync.Mutex
}

// NewWindow 创建指定桶数量、指定统计时长的滑动窗口。
// 桶数量 size 必须大于0，否则 panic。
// 统计时长 time 必须可以被均分为桶数量 size 个，否则 panic。
func NewWindow(size int, duration time.Duration) *Window {
	if size <= 0 {
		panic("The size of rolling window must be greater than 0")
	}
	if int64(duration)%int64(size) != 0 {
		panic(errors.Errorf(
			"The duration of rolling window must divide equally into size: duration: %s / size: %d",
			duration, size))
	}
	w := &Window{
		ring:       NewBucketRing(size + 1),
		size:       size,
		duration:   duration,
		internal:   time.Duration(duration.Nanoseconds() / int64(size)),
		bucketTime: nowFn(),
	}
	// left 到 ring.current 是窗口。
	w.left = w.ring.current
	for i := 0; i < w.size-1; i++ {
		w.ring.Rotate()
	}
	return w
}

func (w *Window) shouldMove() bool {
	return nowFn().Sub(w.bucketTime) >= w.internal
}

func (w *Window) move() {
	w.mu.Lock()
	defer w.mu.Unlock()
	buckets := int(nowFn().Sub(w.bucketTime) / w.internal)
	for i := 0; i < buckets; i++ {
		w.ring.Rotate()
		w.left = w.left.next
		w.bucketTime = w.bucketTime.Add(w.internal)
	}
}

// RecordFunc 用于变更桶中数据的函数
type RecordFunc func(b *Bucket)

// Record 使用 rf 函数变更窗口头部桶内计数
func (w *Window) Record(rf RecordFunc) {
	if w.shouldMove() {
		w.move()
	}
	rf(w.ring.Current())
}

type calculateFunc func(b *Bucket)

func (w *Window) calculate(cf calculateFunc) {
	if w.shouldMove() {
		w.move()
	}
	p := w.left
	for p != w.ring.Current().next {
		cf(p)
		p = p.next
	}
}

// SumSuccess 获取窗口内成功总计数
func (w *Window) SumSuccess() int64 {
	var sum int64
	w.calculate(func(b *Bucket) {
		sum += b.Success
	})
	return sum
}

// SumFailure 获取窗口内失败总计数
func (w *Window) SumFailure() int64 {
	var sum int64
	w.calculate(func(b *Bucket) {
		sum += b.Failure
	})
	return sum
}
