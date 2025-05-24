package redis

var cache = make(map[string]string)

func Set(key, value string) {
	cache[key] = value
}

func Get(key string) (string, bool) {
	value, exists := cache[key]
	return value, exists
}
