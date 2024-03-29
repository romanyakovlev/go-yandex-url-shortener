package repository

type URLRepository interface {
	Save(randomPath string, urlStr string)
	Find(shortURL string) (string, bool)
}

type MemoryURLRepository struct {
	URLMap map[string]string
}

func (r MemoryURLRepository) Save(randomPath string, urlStr string) {
	r.URLMap[randomPath] = urlStr
}

func (r MemoryURLRepository) Find(shortURL string) (string, bool) {
	value, ok := r.URLMap[shortURL]
	return value, ok
}
