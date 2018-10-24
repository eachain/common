package dict

// NewSimpleMap返回普通的map[interface{}]interface{}
func NewSimpleMap() Map {
	return simpleMap(make(map[interface{}]interface{}))
}

type simpleMap map[interface{}]interface{}

func (sm simpleMap) Load(key interface{}) (value interface{}, ok bool) {
	value, ok = sm[key]
	return
}

func (sm simpleMap) LoadOrStore(key interface{}, new func() interface{}) (actual interface{}, loaded bool) {
	actual, loaded = sm[key]
	if loaded {
		return
	}
	actual = new()
	sm[key] = actual
	return
}

func (sm simpleMap) Store(key, value interface{}) {
	sm[key] = value
}

func (sm simpleMap) Delete(key interface{}) {
	delete(sm, key)
}

func (sm simpleMap) Range(f func(key, value interface{}) bool) {
	for key, value := range sm {
		if !f(key, value) {
			return
		}
	}
}

func (sm simpleMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(sm)
}

func (sm simpleMap) UnmarshalJSON(p []byte) error {
	return unmarshalJSON(p, sm)
}
