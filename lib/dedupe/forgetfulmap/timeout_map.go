package forgetfulmap

import "time"

type (
	ExpiringMap struct {
		m *ForgetfulSyncMap
	}

	ForgetableItem struct {
		age   time.Time
		value interface{}
	}
)

func (f *ForgetableItem) CanBeForgotten(oldAfter time.Duration) bool {
	// calc the oldest this item can be
	oldest := time.Now().Add(-oldAfter)
	return !f.age.After(oldest)
}

func (em *ExpiringMap) HasKey(key interface{}) bool {
	return em.m.HasKey(key)
}
func (em *ExpiringMap) AddKey(key interface{}) {
	em.AddKey(ForgetableItem{
		age: time.Now(),
	})
}
func (em *ExpiringMap) Load(key interface{}) (interface{}, bool) {
	fi, found := em.m.Load(key)
	if !found {
		return nil, found
	}
	return fi.(ForgetableItem).value, true
}
func (em *ExpiringMap) Store(key, value interface{}) {
	em.m.Store(key, ForgetableItem{
		age:   time.Now(),
		value: value,
	})
}
func (em *ExpiringMap) Delete(key interface{}) {
	em.m.Delete(key)
}

func (em *ExpiringMap) Len() int32 {
	return em.m.Len()
}

func (em *ExpiringMap) Range(rangeFunc func(key, value interface{}) bool) {
	em.m.Range(func(key, value interface{}) bool {
		return rangeFunc(key, value.(ForgetableItem).value)
	})
}

func (em *ExpiringMap) Stop() {
	em.m.Stop()
}
