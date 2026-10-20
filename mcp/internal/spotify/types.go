package spotify

type Hit struct {
	Index  int
	Title  string
	Artist string
	URI    string
}

type Bar struct {
	Title  string
	Artist string
	URI    string
}

type Device struct {
	ID   string
	Name string
	Type string
}
