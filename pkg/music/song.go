package music

type Song interface {
	GetName() string
	getPath() string
}

type song struct {
	Name string
	path string
}

func NewSong(name string, path string) Song {
	return &song{Name: name, path: path}
}

func (s *song) GetName() string {
	return s.Name
}

func (s *song) getPath() string {
	return s.path
}

type SongInfo struct {
	Artist   string
	Title    string
	Playlist string
	Picture  SongPicture
}

type SongPicture struct {
	Data []byte
	Mime string
}
