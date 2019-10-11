package connection

type audioItem struct {
	opusData chan []byte
	dead     bool // set to true if needs to be cleaned up
}

func (a *audioItem) OpusChan() chan []byte {
	return a.opusData
}

func (a *audioItem) IsClosed() bool {
	return a.dead
}

func (a *audioItem) Cleanup() {
	a.dead = true
	<-a.opusData
}
