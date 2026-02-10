package or_test

import (
	"context"
	"testing"
	"time"

	or "github.com/UnendingLoop/OR-channel-union"
)

func TestOrPackage(t *testing.T) {
	chCases := []struct {
		name         string
		setupChArray func() []<-chan any
		wantTimeMin  time.Duration
		wantTimeMax  time.Duration
	}{
		{
			name:         "empty slice of CH",
			setupChArray: makeSignals(false, 0),
			wantTimeMin:  0 * time.Second,
			wantTimeMax:  900 * time.Millisecond,
		},
		{
			name:         "1 nil-CH in slice",
			setupChArray: makeSignals(true, 1),
			wantTimeMin:  0 * time.Second,
			wantTimeMax:  900 * time.Millisecond,
		},
		{
			name:         "1 non-nil CH in slice",
			setupChArray: makeSignals(false, 1),
			wantTimeMin:  900 * time.Millisecond,
			wantTimeMax:  1200 * time.Millisecond,
		},
		{
			name:         "11 CHs in slice",
			setupChArray: makeSignals(true, 11),
			wantTimeMin:  900 * time.Millisecond,
			wantTimeMax:  1200 * time.Millisecond,
		},
		{
			name:         "5 nil-CHs in slice",
			setupChArray: func() []<-chan any { return []<-chan any{nil, nil, nil, nil, nil} },
			wantTimeMin:  0 * time.Second,
			wantTimeMax:  900 * time.Millisecond,
		},
	}

	t.Run("Testing 'Or' function", func(t *testing.T) {
		t.Parallel()
		for _, tt := range chCases {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				start := time.Now()
				<-or.Or(tt.setupChArray()...)
				end := time.Since(start)
				if end < tt.wantTimeMin || end > tt.wantTimeMax {
					t.Logf("Or-function finished working after %v instead of min %v and max %v", end, tt.wantTimeMin, tt.wantTimeMax)
					t.Fail()
				}
			})
		}
	})

	t.Run("Testing 'OrContext' function", func(t *testing.T) {
		t.Parallel()
		for _, tt := range chCases {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				start := time.Now()
				<-or.OrContext(context.Background(), tt.setupChArray()...)
				end := time.Since(start)
				if end < tt.wantTimeMin || end > tt.wantTimeMax {
					t.Logf("OrContext-function finished working after %v instead of min %v and max %v", end, tt.wantTimeMin, tt.wantTimeMax)
					t.Fail()
				}
			})
		}
	})

	t.Run("Testing 'OrContext' function with cancelled context", func(t *testing.T) {
		t.Parallel()
		for _, tt := range chCases {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				start := time.Now()
				<-or.OrContext(ctx, tt.setupChArray()...)
				end := time.Since(start)
				if end > 200*time.Millisecond {
					t.Logf("OrContext-function finished working after %v instead of approx. 0 second", end)
					t.Fail()
				}
			})
		}
	})
}

func makeSignals(firstChNil bool, len int) func() []<-chan any {
	res := []<-chan any{
		nil,
		chanProducer(1 * time.Second),
		nil,
		chanProducer(5 * time.Second),
		nil,
		chanProducer(10 * time.Second),
		nil,
		chanProducer(15 * time.Second),
		nil,
		chanProducer(20 * time.Second),
		nil,
	}

	switch len {
	case 0:
		return func() []<-chan any { return res[0:0] }
	case 1:
		if firstChNil {
			return func() []<-chan any { return res[0:1] }
		} else {
			return func() []<-chan any { return res[1:2] }
		}
	default:
		if firstChNil {
			return func() []<-chan any { return res[0:] }
		} else {
			return func() []<-chan any { return res[1:] }
		}
	}
}

func chanProducer(after time.Duration) chan any {
	c := make(chan any)
	go func() {
		defer close(c)
		time.Sleep(after)
	}()
	return c
}
