package ws

type Room struct {
	Players    map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
	Shutdown   chan struct{}
}

func NewRoom() *Room {
	return &Room{
		Players:    make(map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Shutdown:   make(chan struct{}),
	}
}

func (r *Room) Run() {
	for {
		select {
		case <-r.Shutdown:
			return
		case player := <-r.Register:
			r.Players[player] = true
		case player := <-r.Unregister:
			delete(r.Players, player)
		}
	}
}
