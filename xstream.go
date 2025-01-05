// Copyright 2025 DongWoo Lee. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
//
// https://github.com/leepro/jstream

// Package jstream provides data stream through websocket

package xstream

import (
	"context"
	"fmt"
	"sync"
)

const (
	MAX_BUFFER = 1024
)

type StreamGroup struct {
	ctx     context.Context
	streams map[string]*Stream
	lock    sync.RWMutex
}

func New(ctx context.Context) *StreamGroup {
	return &StreamGroup{
		ctx:     ctx,
		streams: make(map[string]*Stream),
	}
}

func (sg *StreamGroup) RegisterProducer(ctx context.Context) *Stream {
	sg.lock.Lock()
	defer sg.lock.Unlock()

	sm := NewStream(ctx)
	go sm.Run()

	sg.streams[sm.ID] = sm
	return sm
}

func (sg *StreamGroup) UnregisterProducer(streamID string) error {
	sg.lock.Lock()
	defer sg.lock.Unlock()

	stream, exists := sg.streams[streamID]
	if !exists {
		return fmt.Errorf("stream with ID %s does not exist", streamID)
	}

	stream.Close()

	delete(sg.streams, streamID)

	return nil
}

func (sg *StreamGroup) Close() error {
	sg.lock.Lock()
	defer sg.lock.Unlock()

	for _, stream := range sg.streams {
		stream.lock.Lock()
		close(stream.WriteC)
		for _, sub := range stream.Reads {
			sub.Close()
		}
		stream.lock.Unlock()
	}
	return nil
}

func (sg *StreamGroup) AllSessions() []string {
	sg.lock.RLock()
	defer sg.lock.RUnlock()

	keys := make([]string, 0, len(sg.streams))
	for key := range sg.streams {
		keys = append(keys, key)
	}

	return keys
}

func (sg *StreamGroup) Subscribe(ctx context.Context, streamID string) (*Reader, error) {
	sg.lock.RLock()
	defer sg.lock.RUnlock()

	stream, exists := sg.streams[streamID]
	if !exists {
		return nil, fmt.Errorf("stream with ID %s does not exist", streamID)
	}

	return stream.AddReader(ctx), nil
}

func (sg *StreamGroup) LookupStream(streamID string) (*Stream, error) {
	sg.lock.RLock()
	defer sg.lock.RUnlock()

	stream, exists := sg.streams[streamID]
	if !exists {
		return nil, fmt.Errorf("stream with ID %s does not exist", streamID)
	}

	return stream, nil
}

func (sg *StreamGroup) Unsubscribe(r *Reader) error {
	sm, err := sg.LookupStream(r.StreamID)
	if err != nil {
		return err
	}

	sm.RemoveReader(r.ID)
	return nil
}
