package forgetfulmap

import (
	"sync"
	"time"
)

type (
	ForgetfulSyncMap struct {
		lookup        *sync.Map
		sweeper       *time.Timer
		sweepInterval time.Duration
		oldAfter      time.Duration
		evictionFunc  func(key interface{}, value interface{})
	}

	// ForgettableItem allows us to determine if something can be forgotten
	ForgettableItem interface {
		CanBeForgotten(oldAfter time.Duration) bool
	}

	// a generic wrapper for things that can be lost
	marble struct {
		age   time.Time
		value interface{}
	}
)

func NewForgetfulSyncMap(interval time.Duration, oldTime time.Duration) *ForgetfulSyncMap {
	f := ForgetfulSyncMap{
		lookup:        &sync.Map{},
		sweepInterval: interval,
		oldAfter:      oldTime,
	}
	f.sweeper = time.AfterFunc(f.oldAfter, func() {
		f.sweep()
		f.sweeper.Reset(f.sweepInterval)
	})

	return &f
}

func (f *ForgetfulSyncMap) SetEvictionAction(evictFunc func(key interface{}, value interface{})) {
	f.evictionFunc = evictFunc
}

func (f *ForgetfulSyncMap) sweep() {
	f.lookup.Range(func(key, value interface{}) bool {
		t, ok := value.(ForgettableItem)
		if !ok {
			// cannot forget something which cannot be forgotten
			return true
		}

		if t.CanBeForgotten(f.oldAfter) {
			if f.evictionFunc != nil {
				f.evictionFunc(key, value)
			}
			f.Delete(key)
		}

		return true
	})
}

func (f *ForgetfulSyncMap) HasKey(key interface{}) bool {
	if _, ok := f.lookup.Load(key); ok {
		return true
	}
	return false
}

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
	// TODO: this line is wrong. What does it mean to Add a key without a value?
	f.Store(key, ForgetableItem{})
}

func (f *ForgetfulSyncMap) Load(key interface{}) (interface{}, bool) {
	return f.lookup.Load(key)
}

func (f *ForgetfulSyncMap) Store(key interface{}, value ForgetableItem) {
	f.lookup.Store(key, value)
}

func (f *ForgetfulSyncMap) Delete(key interface{}) {
	f.lookup.Delete(key)
}

func (f *ForgetfulSyncMap) Len() (entries int32) {
	f.lookup.Range(func(key interface{}, value interface{}) bool {
		entries++
		return true
	})

	return entries
}

func (f *ForgetfulSyncMap) Range(rangeFunc func(key, value interface{}) bool) {
	f.lookup.Range(rangeFunc)
}

func (f *ForgetfulSyncMap) Stop() {
	f.sweeper.Stop()
}
