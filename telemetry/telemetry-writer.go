package telemetry

import (
	"context"
	"errors"
	"github.com/anacrolix/chansync"
	"github.com/anacrolix/chansync/events"
	"github.com/anacrolix/log"
	"io"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"slices"
	"strings"
	"sync"
	"time"
)

type Writer struct {
	// websocket and HTTP post are supported. Posting isn't very nice through Cloudflare.
	Url *url.URL
	// Logger for *this*. Probably don't want to loop it back to itself.
	Logger log.Logger
	// The time between reconnects to the Url.
	RetryInterval time.Duration

	// Lazy init guard.
	init   sync.Once
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	mu sync.Mutex
	// This lets loggers not block.
	buf        chan []byte
	retry      [][]byte
	addPending chansync.BroadcastCond

	closed      chansync.SetOnce
	closeReason string
}

func (me *Writer) writerWaitCond() (
	stop bool, // Stop writing
	ready bool, // There are messages ready to go.
	newMessages events.Signaled, // An event for new messages.
) {
	me.mu.Lock()
	defer me.mu.Unlock()
	if me.ctx.Err() != nil {
		// Closed and hard limit.
		stop = true
		return
	}
	if len(me.buf) != 0 || len(me.retry) != 0 {
		ready = true
		return
	}
	if me.closed.IsSet() {
		// We're requested to stop and there's nothing to send.
		stop = true
		return
	}
	// Return the cond chan for new messages.
	newMessages = me.addPending.Signaled()
	return
}

// Returns true if there are messages pending, and false if we should stop writing.
func (me *Writer) writerWait() (ready bool) {
	for {
		stop, ready_, newMessages := me.writerWaitCond()
		if stop {
			return false
		}
		if ready_ {
			return true
		}
		select {
		case <-newMessages:
		case <-me.closed.Done():
		case <-me.ctx.Done():
		}
	}
}

func (me *Writer) writer() {
	defer me.wg.Done()
	for {
		if !me.writerWait() {
			return
		}
		me.Logger.Levelf(log.Debug, "connecting")
		wait := func() bool {
			if strings.Contains(me.Url.Scheme, "ws") {
				return me.websocket()
			} else {
				me.streamPost()
				return true
			}
		}()
		if wait && me.closed.IsSet() {
			// We just failed, and have been closed. Don't try again.
			return
		}
		if wait {
			select {
			case <-time.After(me.RetryInterval):
			case <-me.closed.Done():
			}
		}
	}
}

// Waits a while to allow final messages to go through. Another method should be added to make this
// customizable. Nothing should be logged after calling this.
func (me *Writer) Close(reason string) error {
	me.lazyInit()
	me.closeReason = reason
	me.closed.Set()
	me.Logger.Levelf(log.Debug, "waiting for writer")
	close(me.buf)
	go func() {
		time.Sleep(5 * time.Second)
		me.cancel()
	}()
	me.wg.Wait()
	return nil
}

// wait is true if the caller should wait a while before retrying.
func (me *Writer) websocket() (wait bool) {
	conn, _, err := websocket.Dial(me.ctx, me.Url.String(), nil)
	if err != nil {
		me.Logger.Levelf(log.Error, "error dialing websocket: %v", err)
		return true
	}
	defer func() {
		err := context.Cause(me.ctx)
		reason := me.closeReason
		if err != nil {
			reason = err.Error()
		}
		conn.Close(websocket.StatusNormalClosure, reason)
	}()
	ctx, cancel := context.WithCancel(me.ctx)
	go func() {
		err := me.payloadWriter(
			ctx,
			func(b []byte) error {
				err := conn.Write(ctx, websocket.MessageText, b)
				me.Logger.Levelf(log.Debug, "wrote %q: %v", b, err)
				return err
			},
		)
		if err != nil {
			me.Logger.Levelf(log.Error, "payload writer failed: %v", err)
		}
		// Notify that we're not sending anymore.
		err = conn.Write(ctx, websocket.MessageBinary, nil)
		if err != nil {
			me.Logger.Levelf(log.Error, "writing end of stream: %v", err)
		}
	}()
	err = me.websocketReader(me.ctx, conn)
	// Since we can't receive acks anymore, stop sending immediately.
	cancel()
	me.Logger.Levelf(log.Error, "reading from websocket: %v", err)
	return false
}

func (me *Writer) websocketReader(ctx context.Context, conn *websocket.Conn) error {
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		me.Logger.Levelf(log.Debug, "read from telemetry websocket: %q", string(data))
	}
}

func (me *Writer) streamPost() {
	ctx, cancel := context.WithCancel(me.ctx)
	defer cancel()
	r, w := io.Pipe()
	go func() {
		defer w.Close()
		err := me.payloadWriter(ctx, func(b []byte) error {
			_, err := w.Write(b)
			return err
		})
		if err != nil {
			me.Logger.Levelf(log.Error, "http post payload writer failed: %v", err)
		}
	}()
	me.Logger.Levelf(log.Debug, "starting post")
	// What's the content type for newline/ND/packed JSON streams?
	resp, err := http.Post(me.Url.String(), "application/jsonl", r)
	me.Logger.Levelf(log.Debug, "post returned")
	r.Close()
	if err != nil {
		me.Logger.Levelf(log.Error, "error posting: %s", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		me.Logger.Levelf(log.Error, "unexpected status code: %v", resp.StatusCode)
	}
	resp.Body.Close()
}

func (me *Writer) payloadWriter(ctx context.Context, w func(b []byte) error) error {
	for {
		select {
		case b, ok := <-me.buf:
			if !ok {
				me.Logger.Levelf(log.Debug, "buf closed")
				return nil
			}
			me.Logger.Levelf(log.Debug, "writing %v byte payload", len(b))
			err := w(b)
			if err != nil {
				me.Logger.Levelf(log.Debug, "error writing payload: %s", err)
				me.retry = append(me.retry, b)
				me.addPending.Broadcast()
				return err
			}
		case <-ctx.Done():
			return context.Cause(me.ctx)
		}
	}
}

func (me *Writer) lazyInit() {
	me.init.Do(func() {
		if me.Logger.IsZero() {
			me.Logger = log.Default
		}
		me.buf = make(chan []byte, 256)
		me.ctx, me.cancel = context.WithCancel(context.Background())
		if me.RetryInterval == 0 {
			me.RetryInterval = time.Minute
		}
		me.wg.Add(1)
		go me.writer()
	})
}

func (me *Writer) Write(p []byte) (n int, err error) {
	me.lazyInit()
	select {
	// Wow, thanks for not reporting this with race detector, Go.
	case me.buf <- slices.Clone(p):
		me.addPending.Broadcast()
		return len(p), nil
	default:
		me.Logger.Levelf(log.Error, "payload lost")
		return 0, errors.New("payload lost")
	}
}
