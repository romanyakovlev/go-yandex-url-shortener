package repository

type MemoryURLRepository struct {
	URLMap map[string]string
}

func (r MemoryURLRepository) Save(randomPath string, urlStr string) error {
	r.URLMap[randomPath] = urlStr
	return nil
}

func (r MemoryURLRepository) Find(shortURL string) (string, bool) {
	value, ok := r.URLMap[shortURL]
	return value, ok
}
