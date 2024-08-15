package socks

import (
	"context"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
)

type Proxy struct {
	mu   sync.Mutex
	done atomic.Value
	err  error

	lconn, rconn io.ReadWriteCloser
}

func NewProxy(lconn, rconn io.ReadWriteCloser) *Proxy {
	return &Proxy{
		lconn: lconn,
		rconn: rconn,
	}
}

func (c *Proxy) Start(ctx context.Context) {

	//bidirectional copy
	go c.pipe(ctx, c.lconn, c.rconn)
	go c.pipe(ctx, c.rconn, c.lconn)

	//wait for close...
	<-c.Done()
}

func (c *Proxy) pipe(ctx context.Context, src, dst io.ReadWriter) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.Done():
			return nil
		default:
		}

		_, err := io.Copy(dst, src)
		if err != nil {
			c.Cancel(err)

			return fmt.Errorf("io.Copy: %w", err)
		}
	}
}

func (c *Proxy) Done() <-chan struct{} {
	d := c.done.Load()
	if d != nil {
		return d.(chan struct{})
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	d = c.done.Load()
	if d == nil {
		d = make(chan struct{})
		c.done.Store(d)
	}
	return d.(chan struct{})
}

func (c *Proxy) Err() error {
	c.mu.Lock()
	err := c.err
	c.mu.Unlock()
	return err
}

// closedchan is a reusable closed channel.
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

func (c *Proxy) Cancel(err error) {
	c.mu.Lock()
	if c.err != nil {
		c.mu.Unlock()
		return // already canceled
	}
	c.err = err

	d, _ := c.done.Load().(chan struct{})
	if d == nil {
		c.done.Store(closedchan)
	} else {
		close(d)
	}

	c.mu.Unlock()
}
