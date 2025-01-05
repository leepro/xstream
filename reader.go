// Copyright 2025 DongWoo Lee. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// https://github.com/leepro/jstream

// Package jstream provides data stream through websocket

package xstream

import (
	"context"
	"sync"
)

type Reader struct {
	ID       string
	StreamID string
	C        chan any

	ctx      context.Context
	isClosed bool
	lock     sync.RWMutex
}

// Subscriber
func (r *Reader) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.isClosed {
		return nil
	}

	r.isClosed = true
	close(r.C)
	return nil
}
