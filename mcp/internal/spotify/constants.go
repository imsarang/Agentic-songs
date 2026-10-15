package spotify

const WebPlayerURL = "https://open.spotify.com"

type View string

const (
	ViewPlayer  View = "player"
	ViewLogin   View = "login"
	ViewUnknown View = "unknown"
)
