package telemetry

import (
	"context"
	"errors"
	"github.com/anacrolix/chansync"
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
	init sync.Once
	// This lets loggers not block.
	buf    chan []byte
	retry  [][]byte
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	closed      chansync.SetOnce
	closeReason string
}

func (me *Writer) writer() {
	defer me.wg.Done()
	for {
		if me.closed.IsSet() && len(me.buf) == 0 && len(me.retry) == 0 {
			return
		}
		select {
		case <-me.ctx.Done():
			return
		default:
		}
		wait := func() bool {
			if strings.Contains(me.Url.Scheme, "ws") {
				return me.websocket()
			} else {
				me.streamPost()
				return true
			}
		}()
		if me.ctx.Err() != nil {
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
		err := me.ctx.Err()
		reason := me.closeReason
		if err != nil {
			reason = err.Error()
		}
		conn.Close(websocket.StatusNormalClosure, reason)
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		me.payloadWriter(func(b []byte) error {
			err := conn.Write(me.ctx, websocket.MessageText, b)
			me.Logger.Levelf(log.Debug, "wrote %q: %v", b, err)
			return err
		})
	}()
	wg.Wait()
	return false
}

func (me *Writer) streamPost() {
	r, w := io.Pipe()
	go me.payloadWriter(func(b []byte) error {
		_, err := w.Write(b)
		return err
	})
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

func (me *Writer) payloadWriter(w func(b []byte) error) {
	for {
		select {
		case b, ok := <-me.buf:
			if !ok {
				return
			}
			me.Logger.Levelf(log.Debug, "writing %v byte payload", len(b))
			err := w(b)
			if err != nil {
				me.Logger.Levelf(log.Debug, "error writing payload: %s", err)
				me.retry = append(me.retry, b)
				return
			}
		case <-me.ctx.Done():
			return
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
		return len(p), nil
	default:
		me.Logger.Levelf(log.Error, "payload lost")
		return 0, errors.New("payload lost")
	}
}
