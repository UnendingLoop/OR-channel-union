// Package or provides utility for uniting input done-channels in 2 ways:
//
// - classic OR-channel,
//
// - experimental/auxiliary solution: classic OR-channel which respects
// externally provided context.
//
// Both functions:
//
// - receive a number of done-channels and return an unbuffered result-channel
// (available only for reading), which is closed when one of
// the input channels is closed(or becomes readable, which for
// done-channels indicates completion),
//
// - have nil-channel-protection - all nil-input-channels are excluded
// from further listening,
//
// - are non-goroutine-leaky.
package or

import (
	"context"
	"sync"
)

// Or function receives a list of done-channels as slice of channels and
// returns resulting channel, which closes if any non-nil input done-channel
// is closed or receives a value. Every non-nil input channel is given to
// a separate goroutine to be listened.
//
// Pool of goroutines is synced through internal context: when some input
// channel is closed, its listening goroutine closes resulting channel and
// cancels the internal context, which leads to termination of goroutines
// remaining in the pool.
//
// Usage example:
//
//	done := or.Or(ch1, ch2, ch3...)
//	<-done
//
// After 'done' is closed, you may safely finish the work of your app - or not:
// it is up to you to decide depending on your business-rules.
func Or(channels ...<-chan any) <-chan struct{} {
	out := make(chan struct{})

	// готовим контекст для синхронизации горутин
	ctx, cancel := context.WithCancel(context.Background())

	// отсеиваем ниловые каналы
	filtered := make([]<-chan any, 0)
	for _, v := range channels {
		if v == nil {
			continue
		}
		filtered = append(filtered, v)
	}

	// готовим 'одноразовую' функцию для закрытия результирующего канала
	closeOnce := sync.Once{}

	// запускаем цикл
	switch len(filtered) {
	case 0:
		cancel()
		close(out)
		return out
	case 1:
		cancel()
		go func() {
			<-filtered[0]
			close(out)
		}()
		return out
	default:
		for _, v := range filtered {
			go func(ch <-chan any) {
				select {
				case <-ctx.Done():
					return
				case <-ch:
					cancel()
					closeOnce.Do(func() { close(out) })
				}
			}(v)
		}
	}

	// возврат результирующего канала
	return out
}

// OrContext function receives:
//
// - a parent-context,
//
// - a list of done-channels as slice of channels,
//
// and returns resulting channel, which closes if any non-nil input done-channel
// is closed or receives a value. Every non-nil input channel is given to
// a separate goroutine to be listened.
//
// Pool of goroutines is synced through internal context, created using a parent-context:
// when some input done-channel is closedwhich is closed when any of the input channels is
// closed(or becomes readable, which for done-channels indicates completion), its listening
// goroutine closes resulting channel and cancels the internal context, which leads to
// termination of goroutines remaining in the pool.
//
// In case if the parent-context is cancelled before any input done-channel is closed,
// the whole goroutine-pool finishes their work, the resulting channel is also gets closed.
//
// Usage example:
//
//	done := or.OrContext(parentCTX, ch1, ch2, ch3...)
//	<-done
//
// After 'done' is closed, you may safely finish the work of your app - or not:
// it is up to you to decide depending on your business-rules.
func OrContext(ctx context.Context, channels ...<-chan any) <-chan struct{} {
	out := make(chan struct{})

	// отсеиваем ниловые каналы
	filtered := make([]<-chan any, 0)
	for _, v := range channels {
		if v == nil {
			continue
		}
		filtered = append(filtered, v)
	}

	// готовим 'одноразовую' функцию для закрытия результирующего канала
	closeOnce := sync.Once{}

	// запускаем цикл
	switch len(filtered) {
	case 0:
		close(out)
		return out
	case 1:
		go func() {
			select {
			case <-ctx.Done():
				close(out)
			case <-filtered[0]:
				close(out)
			}
		}()
		return out
	default:
		orCtx, cancel := context.WithCancel(ctx)

		for _, v := range filtered {
			go func(ch <-chan any) {
				select {
				case <-orCtx.Done(): // проверка отмены контекста
					closeOnce.Do(func() { close(out) })
				case <-ch:
					closeOnce.Do(func() {
						close(out)
						cancel()
					})
				}
			}(v)
		}
	}

	// возврат результирующего канала
	return out
}
