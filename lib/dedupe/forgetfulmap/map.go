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

	ForgetableItem struct {
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
	var remove bool

	oldest := time.Now().Add(-f.oldAfter)
	f.lookup.Range(func(key, value interface{}) bool {
		remove = true

		if t, ok := value.(ForgetableItem).age, true; ok {
			if t.After(oldest) {
				remove = false
			}
		}

		if remove {
			if f.evictionFunc != nil {
				f.evictionFunc(key, value)
			}
			f.lookup.Delete(key)
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
	f.lookup.Store(key, ForgetableItem{
		age: time.Now(),
	})
}

func (f *ForgetfulSyncMap) Load(key interface{}) (value interface{}, ok bool) {
	retVal, retBool := f.lookup.Load(key)

	if retVal != nil {
		return retVal.(ForgetableItem).value, retBool
	} else {
		return retVal, retBool
	}
}

func (f *ForgetfulSyncMap) Store(key, value interface{}) {
	f.lookup.Store(key, ForgetableItem{
		age:   time.Now(),
		value: value,
	})
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
