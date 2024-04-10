package repository

type URLRepository interface {
	Save(randomPath string, urlStr string) error
	Find(shortURL string) (string, bool)
}
