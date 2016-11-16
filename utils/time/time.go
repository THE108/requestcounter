package time

import "time"

type ITime interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

type RealTime struct{}

func NewRealTime() *RealTime {
	return &RealTime{}
}

func (rt *RealTime) Now() time.Time {
	return time.Now()
}

func (rt *RealTime) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

type FakeTime struct {
	now   time.Time
	chans []chan time.Time
}

func NewFakeTime() *FakeTime {
	return &FakeTime{}
}

func (ft *FakeTime) SetNow(t time.Time) {
	ft.now = t
}

func (ft *FakeTime) Now() time.Time {
	return ft.now
}

func (ft *FakeTime) After(_ time.Duration) <-chan time.Time {
	ch := make(chan time.Time, 1)
	ft.chans = append(ft.chans, ch)
	return ch
}

func (ft *FakeTime) Notify() {
	for _, ch := range ft.chans {
		ch <- ft.now
	}
}
