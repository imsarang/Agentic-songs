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

// Playlist is enough to find a list and add a track.
type Playlist struct {
	ID          string
	Name        string
	Description string
	URI         string
}

type TrackFeatures struct {
	ID            string  `json:"id"`
	Acousticness  float32 `json:"acousticness"`
	Danceability  float32 `json:"danceability"`
	Energy        float32 `json:"energy"`
	Key           int32   `json:"key"`
	Liveness      float32 `json:"liveness"`
	Loudness      float32 `json:"loudness"`
	Speechiness   float32 `json:"speechiness"`
	Mode          int32   `json:"mode"`
	Tempo         float32 `json:"tempo"`
	TimeSignature float32 `json:"time_signature"`
	Valence       float32 `json:"valence"`
}
