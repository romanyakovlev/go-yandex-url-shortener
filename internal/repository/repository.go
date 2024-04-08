package repository

type URLRepository interface {
	Save(randomPath string, urlStr string)
	Find(shortURL string) (string, bool)
}
