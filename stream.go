// Copyright 2025 DongWoo Lee. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// https://github.com/leepro/jstream

// Package jstream provides data stream through websocket

package xstream

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Stream struct {
	ID     string
	WriteC chan any
	Reads  map[string]*Reader

	ctx    context.Context
	cancel context.CancelFunc
	lock   sync.RWMutex
}

func NewStream(ctx context.Context) *Stream {
	ctx, cancel := context.WithCancel(ctx)
	return &Stream{
		ctx:    ctx,
		cancel: cancel,
		ID:     uuid.New().String(),
		WriteC: make(chan any, MAX_BUFFER),
		Reads:  make(map[string]*Reader),
	}
}

func (sm *Stream) AddReader(ctx context.Context) *Reader {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	sub := &Reader{
		ctx:      ctx,
		ID:       uuid.New().String(),
		StreamID: sm.ID,
		C:        make(chan any, MAX_BUFFER),
	}

	sm.Reads[sub.ID] = sub
	return sub
}

func (sm *Stream) Run() {
	for d := range sm.WriteC {

		select {
		case <-sm.ctx.Done():
			log.Printf("Stream %s context done", sm.ID)
			return
		default:
		}

		sm.broadcast(d)
		log.Printf("Stream %s received data %v", sm.ID, d)
	}
}

func (sm *Stream) broadcast(i any) {
	sm.lock.RLock()
	defer sm.lock.RUnlock()

	for _, sub := range sm.Reads {
		sub.C <- i
	}
}

func (sm *Stream) CloseSubscribers() {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	if len(sm.Reads) > 0 {
		for {
			if len(sm.WriteC) == 0 {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}

	for _, sub := range sm.Reads {
		log.Printf("Stream %s closing subscriber %s", sm.ID, sub.ID)
		sub.Close()
	}
}

func (sm *Stream) Close() {
	sm.lock.Lock()
	defer sm.lock.Unlock()

	close(sm.WriteC)
	for _, sub := range sm.Reads {
		sub.Close()
	}

	sm.cancel()
}
