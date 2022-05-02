package forgetfulmap

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type (
	ForgettableFunc func(key, value any, added time.Time) bool
	EvictionFunc    func(key, value any)

	ForgetfulSyncMap struct {
		lookup        *sync.Map
		sweeper       *time.Timer
		sweepInterval time.Duration
		oldAfter      time.Duration
		evictionFunc  EvictionFunc
		forgettable   ForgettableFunc

		itemCounter prometheus.Gauge
	}

	// a generic wrapper for things that can be lost
	marble struct {
		added time.Time
		value any
	}

	Option func(*ForgetfulSyncMap)
)

func NewForgetfulSyncMap(opts ...Option) *ForgetfulSyncMap {
	f := &ForgetfulSyncMap{
		lookup:        &sync.Map{},
		sweepInterval: 10 * time.Second,
		oldAfter:      60 * time.Second,
	}
	for _, opt := range opts {
		opt(f)
	}
	f.sweeper = time.AfterFunc(f.sweepInterval, func() {
		f.sweep()
		f.sweeper.Reset(f.sweepInterval)
	})
	if nil == f.forgettable {
		f.forgettable = OldAfterForgettableAction(f.oldAfter)
	}

	return f
}

func WithPrometheusCounters(numItems prometheus.Gauge) Option {
	return func(syncMap *ForgetfulSyncMap) {
		syncMap.itemCounter = numItems
	}
}

// WithSweepIntervalSeconds is a helper function that Sets the Sweep Interval, so you don't have to cast as much
func WithSweepIntervalSeconds(numSeconds int) Option {
	return WithSweepInterval(time.Duration(numSeconds) * time.Second)
}

// WithOldAgeAfterSeconds is a helper function that Sets the Old Age, so you don't have to cast as much
func WithOldAgeAfterSeconds(numSeconds int) Option {
	return WithOldAgeAfter(time.Duration(numSeconds) * time.Second)
}

// WithSweepInterval sets how often we look to expire items
func WithSweepInterval(d time.Duration) Option {
	return func(f *ForgetfulSyncMap) {
		f.sweepInterval = d
	}
}

// WithOldAgeAfter sets the time at which our default forgetting action declares an items should be forgotten
func WithOldAgeAfter(d time.Duration) Option {
	return func(f *ForgetfulSyncMap) {
		f.oldAfter = d
	}
}

// OldAfterForgettableAction is the default action to determine if an item in our sync.Map should be forgotten.
// it is simply "removable after being in the sync.Map for the given time.Duration"
func OldAfterForgettableAction(oldAfter time.Duration) ForgettableFunc {
	return func(key any, value any, added time.Time) bool {
		oldest := time.Now().Add(-oldAfter)
		return added.Before(oldest)
	}
}

// WithPreEvictionAction Sets the function that is called just before we evict the item
func WithPreEvictionAction(evictFunc EvictionFunc) Option {
	return func(f *ForgetfulSyncMap) {
		f.evictionFunc = evictFunc
	}
}

// WithForgettableAction sets the action that determines if an item should be forgotten
// if your ForgettableFunc returns true, the eviction action is called and then the item is removed
func WithForgettableAction(forgettable ForgettableFunc) Option {
	return func(f *ForgetfulSyncMap) {
		f.forgettable = forgettable
	}
}

// sweep periodically goes through the underlying sync.Map and removes things that should be forgotten
func (f *ForgetfulSyncMap) sweep() {
	if log.Trace().Enabled() {
		log.Trace().Str("section", "forgetfulmap").Int32("num items", f.Len()).Msg("Before Sweep")
	}
	counter := 0
	f.lookup.Range(func(key, value interface{}) bool {
		m, ok := value.(*marble)
		if !ok {
			// not entirely sure how a non marble object got in, but whatever
			return true
		}

		if f.forgettable(key, m.value, m.added) {
			if f.evictionFunc != nil {
				f.evictionFunc(key, value)
			}
			f.Delete(key)
		} else {
			counter++
		}

		return true
	})
	if nil != f.itemCounter {
		f.itemCounter.Set(float64(counter))
	}
	if log.Trace().Enabled() {
		log.Trace().Str("section", "forgetfulmap").Int32("num items", f.Len()).Msg("After Sweep")
	}
}

// HasKey determines if we can Recall an item
// You should use the Load method if you need the return value
func (f *ForgetfulSyncMap) HasKey(key interface{}) bool {
	if _, ok := f.lookup.Load(key); ok {
		return true
	}
	return false
}

// AddKey adds an item to the list without a value
func (f *ForgetfulSyncMap) AddKey(key interface{}) {
	// avoid storing empty things
	if nil == key {
		return
	}
	if kb, ok := key.([]byte); ok {
		if 0 == len(kb) {
			return
		}
	}
	if ks, ok := key.(string); ok {
		if "" == ks {
			return
		}
	}
	f.Store(key, nil)
}

// Load attempts to recall an item from the list
func (f *ForgetfulSyncMap) Load(key any) (any, bool) {
	retVal, retBool := f.lookup.Load(key)

	if retBool {
		if t, tok := retVal.(*marble); tok {
			return t.value, retBool
		} else {
			return nil, false
		}
	} else {
		return retVal, retBool
	}
}

// Store remembers an item
func (f *ForgetfulSyncMap) Store(key, value interface{}) {
	f.lookup.Store(key, &marble{
		added: time.Now(),
		value: value,
	})
}

// Delete Removes an item from the list
func (f *ForgetfulSyncMap) Delete(key interface{}) {
	f.lookup.Delete(key)
}

// Len returns a count of the number of items in the list
func (f *ForgetfulSyncMap) Len() (entries int32) {
	f.lookup.Range(func(key interface{}, value interface{}) bool {
		entries++
		return true
	})

	return entries
}

// Range Iterates over the underlying sync.Map and calls the user function once per item
func (f *ForgetfulSyncMap) Range(rangeFunc func(key, value interface{}) bool) {
	f.lookup.Range(func(key, value interface{}) bool {
		if m, ok := value.(*marble); ok {
			return rangeFunc(key, m.value)
		} else {
			return rangeFunc(key, value)
		}

	})
}

// Stop tells us to stop forgetting.
func (f *ForgetfulSyncMap) Stop() {
	f.sweeper.Stop()
}
