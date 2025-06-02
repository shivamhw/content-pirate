package store


type Store interface {
	Write(path string, data []byte) error
	DirExists(string) bool
	CreateDir(string)
	CleanAll(string) error
}