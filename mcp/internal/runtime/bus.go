package runtime

// constructs the channels , later tools share
func BusInit() *Bus {
	return &Bus{
		CommandsChannel:      make(chan Command, commandBuffer),
		SessionEventsChannel: make(chan SessionEvent, eventBuffer),
		ReadyChannel:         make(chan struct{}),
		ActorStoppedChannel:  make(chan struct{}),
	}
}

// MarkReady closes Ready once, which broadcasts login success to every waiter.
func (b *Bus) MarkReady() {
	b.readyOnce.Do(func() {
		close(b.ReadyChannel)
	})
}
