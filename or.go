package or

import (
	"sync"
)

func RecursiveChannelUnion(channels ...<-chan interface{}) <-chan interface{} {
	var or func(channels ...<-chan interface{}) <-chan interface{}

	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}
		out := make(chan interface{})
		go func() {
			select {
			case <-channels[0]:
			case <-or(channels[1:]...):
			}
			defer close(out)
		}()

		return out
	}
	return or(channels...)
}

func CyclicChannelUnion(channels ...chan interface{}) chan interface{} {
	out := make(chan interface{})
	closeOnce := sync.Once{}
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	default:
		for _, v := range channels {
			go func(ch chan interface{}) {
				<-ch
				closeOnce.Do(func() { close(out) })
			}(v)
		}
	}
	return out
}
