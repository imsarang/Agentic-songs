package runtime

import (
	"context"
	"sync"
)

// The seven tool kinds are declared now so later phases reuse this type.
type Kind string

// enums
const (
	KindEnsureSession Kind = "ensureSession"
	KindSession       Kind = "session"
	KindSearch        Kind = "search"
	KindPlay          Kind = "play"
	KindControl       Kind = "control"
	KindNowPlaying    Kind = "nowPlaying"
	KindPlaylistAdd   Kind = "playlistAdd"
	KindLyrics        Kind = "lyrics"
)

// Phase is the session state the supervisor remembers
type Phase string

const (
	PhaseStarting  Phase = "starting"
	PhaseLoggingIn Phase = "loggingIn"
	PhaseChallenge Phase = "challenge"
	PhaseReady     Phase = "ready"
	PhaseExpired   Phase = "expired"
)

// Event is a login fact the actor publishes. SuperVisor is the consumer

type Event string

const (
	EventNeedsLogin     Event = "needsLogin"
	EventLoggedIn       Event = "loggedIn"
	EventChallenge      Event = "challenge"
	EventExpired        Event = "expired"
	EventBrowserCrashed Event = "browserCrashed"
)

// Args are the fields later tools fill in. ensureSession leaves them empty.
type Args struct {
	Query        string
	Limit        int
	Action       string
	PlaylistName string
	Title        string
	Artist       string
	ResultIndex  int
	TrackID      string
	UseCurrent   bool
}

// Command is one unit of job/work for the actor
type Command struct {
	Kind         Kind
	Args         Args
	ReplyChannel chan Result
	Ctx          context.Context
}

// Result is the single value sent on a command's reply channel.
type Result struct {
	Phase Phase
	Text  string
	Err   error
}

// SessionEvent is one value on the sessionEvents channel.
type SessionEvent struct {
	Type  Event
	Phase Phase
	Err   error
}

// BUS Conatants
const (
	commandBuffer = 8
	eventBuffer   = 4
)

// Bus is the channel bundle. Buffer sizes live here.
type Bus struct {
	CommandsChannel      chan Command
	SessionEventsChannel chan SessionEvent
	ReadyChannel         chan struct{}
	ActorStoppedChannel  chan struct{}
	readyOnce            sync.Once
}
