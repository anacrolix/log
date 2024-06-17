package telemetry

import (
	"context"
	"errors"
	"github.com/anacrolix/log"
	"io"
	"net/http"
	"net/url"
	"nhooyr.io/websocket"
	"strings"
	"sync"
	"time"
)

type Writer struct {
	Context context.Context
	Url     *url.URL
	Logger  log.Logger
	init    sync.Once
	buf     chan []byte
	retry   [][]byte
}

func (me *Writer) writer() {
	for {
		wait := func() bool {
			if strings.Contains(me.Url.Scheme, "ws") {
				return me.websocket()
			} else {
				me.streamPost()
				return true
			}
		}()
		if me.Context.Err() != nil {
			return
		}
		if wait {
			time.Sleep(time.Minute)
		}
	}
}

func (me *Writer) websocket() (wait bool) {
	conn, _, err := websocket.Dial(me.Context, me.Url.String(), nil)
	if err != nil {
		me.Logger.Levelf(log.Error, "error dialing websocket: %v", err)
		return true
	}
	defer func() {
		conn.Close(websocket.StatusNormalClosure, me.Context.Err().Error())
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		me.payloadWriter(func(b []byte) error {
			return conn.Write(me.Context, websocket.MessageText, b)
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
		case <-me.Context.Done():
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
		go me.writer()
	})
}

func (me *Writer) Write(p []byte) (n int, err error) {
	me.lazyInit()
	select {
	case me.buf <- p:
		return len(p), nil
	default:
		me.Logger.Levelf(log.Error, "payload lost")
		return 0, errors.New("payload lost")
	}
}
