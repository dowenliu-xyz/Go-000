package rolling

import (
	"testing"
	"time"
)

func TestWindow(t *testing.T) {
	testTime := time.Now()
	nowFn = func() time.Time {
		return testTime
	}
	w := NewWindow(3, 3*time.Second)
	w.Record(IncreaseSuccess)
	nowFn = func() time.Time {
		return testTime.Add(time.Second)
	}
	w.Record(IncreaseSuccess)
	w.Record(IncreaseSuccess)
	ss1 := w.SumSuccess()
	if ss1 != 3 {
		t.Errorf("expect window success sum being 3 but %d", ss1)
	}
	nowFn = func() time.Time {
		return testTime.Add(time.Second * 3)
	}
	ss2 := w.SumSuccess()
	if ss2 != 2 {
		t.Errorf("expect window success sum being 2 but %d", ss2)
	}
}
