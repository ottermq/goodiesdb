package server

import (
	"net"
	"path"
	"sync"

	"github.com/ottermq/goodiesdb/internal/logging"
)

const pubsubDeliveryBufSize = 64

// PubSubBroker routes published messages to subscribed connections.
// Channels are global across all databases, matching Redis behaviour.
//
// Each connection gets exactly one delivery channel (connChans). All subscriptions
// for a connection — exact channels and patterns — share that single channel, which
// maps directly to the connection's write goroutine.
type PubSubBroker struct {
	mu        sync.RWMutex
	channels  map[string]map[net.Conn]bool // exact channel → set of subscribed conns
	patterns  map[string]map[net.Conn]bool // glob pattern  → set of subscribed conns
	connChans map[net.Conn]chan []byte     // conn → single delivery queue
}

func newPubSubBroker() *PubSubBroker {
	return &PubSubBroker{
		channels:  make(map[string]map[net.Conn]bool),
		patterns:  make(map[string]map[net.Conn]bool),
		connChans: make(map[net.Conn]chan []byte),
	}
}

// getOrCreateChan returns the delivery channel for conn, creating it if needed.
// Must be called with b.mu held (write lock).
func (b *PubSubBroker) getOrCreateChan(conn net.Conn) chan []byte {
	ch, ok := b.connChans[conn]
	if !ok {
		ch = make(chan []byte, pubsubDeliveryBufSize)
		b.connChans[conn] = ch
	}
	return ch
}

// GetConnChan returns the delivery channel for conn, or nil if the conn has no
// subscriptions. Used by the server to start the write goroutine.
func (b *PubSubBroker) GetConnChan(conn net.Conn) <-chan []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.connChans[conn]
}

// Subscribe registers conn on an exact channel. Returns the total number of
// exact+pattern subscriptions for this conn after the operation.
func (b *PubSubBroker) Subscribe(conn net.Conn, channel string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.getOrCreateChan(conn)

	if b.channels[channel] == nil {
		b.channels[channel] = make(map[net.Conn]bool)
	}
	b.channels[channel][conn] = true

	return b.countSubscriptions(conn)
}

// Unsubscribe removes conn from an exact channel. Returns the remaining total
// subscriptions for this conn.
func (b *PubSubBroker) Unsubscribe(conn net.Conn, channel string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.channels[channel]; ok {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(b.channels, channel)
		}
	}

	remaining := b.countSubscriptions(conn)
	if remaining == 0 {
		b.closeAndRemoveConnChan(conn)
	}
	return remaining
}

// PSubscribe registers conn on a glob pattern. Returns the total subscriptions
// for this conn after the operation.
func (b *PubSubBroker) PSubscribe(conn net.Conn, pattern string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.getOrCreateChan(conn)

	if b.patterns[pattern] == nil {
		b.patterns[pattern] = make(map[net.Conn]bool)
	}
	b.patterns[pattern][conn] = true

	return b.countSubscriptions(conn)
}

// PUnsubscribe removes conn from a glob pattern. Returns the remaining total
// subscriptions for this conn.
func (b *PubSubBroker) PUnsubscribe(conn net.Conn, pattern string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.patterns[pattern]; ok {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(b.patterns, pattern)
		}
	}

	remaining := b.countSubscriptions(conn)
	if remaining == 0 {
		b.closeAndRemoveConnChan(conn)
	}
	return remaining
}

// Publish fans out msg to all exact subscribers of channel and all pattern
// subscribers whose pattern matches channel. Returns the total number of
// connections that received the message.
func (b *PubSubBroker) Publish(channel, msg string) int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	received := 0

	// Exact subscribers
	if subs, ok := b.channels[channel]; ok {
		payload := encodeMessage(channel, msg)
		for conn := range subs {
			if sent := b.deliver(conn, payload); sent {
				received++
			}
		}
	}

	// Pattern subscribers — skip conns already counted via exact match
	exactSubs := b.channels[channel]
	for pattern, subs := range b.patterns {
		matched, err := path.Match(pattern, channel)
		if err != nil || !matched {
			continue
		}
		payload := encodePMessage(pattern, channel, msg)
		for conn := range subs {
			if exactSubs[conn] {
				// conn already received an exact message; still send the pmessage
				// (Redis sends both), but count it only once for the return value.
				b.deliver(conn, payload)
				continue
			}
			if sent := b.deliver(conn, payload); sent {
				received++
			}
		}
	}

	return received
}

// deliver sends payload to the conn's delivery channel. Returns false if the
// buffer was full and the message was dropped. Must be called with b.mu held.
func (b *PubSubBroker) deliver(conn net.Conn, payload []byte) bool {
	ch, ok := b.connChans[conn]
	if !ok {
		return false
	}
	select {
	case ch <- payload:
		return true
	default:
		logging.Errorf("pubsub: delivery buffer full for conn %v, dropping message", conn.RemoteAddr())
		return false
	}
}

// UnsubscribeAll removes conn from every channel and pattern. Called on disconnect.
func (b *PubSubBroker) UnsubscribeAll(conn net.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	for channel, subs := range b.channels {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(b.channels, channel)
		}
	}
	for pattern, subs := range b.patterns {
		delete(subs, conn)
		if len(subs) == 0 {
			delete(b.patterns, pattern)
		}
	}
	b.closeAndRemoveConnChan(conn)
}

// closeAndRemoveConnChan closes and removes the delivery channel for conn.
// Must be called with b.mu held (write lock).
func (b *PubSubBroker) closeAndRemoveConnChan(conn net.Conn) {
	if ch, ok := b.connChans[conn]; ok {
		close(ch)
		delete(b.connChans, conn)
	}
}

// countSubscriptions returns the total number of exact channels and patterns
// this conn is subscribed to. Must be called with b.mu held.
func (b *PubSubBroker) countSubscriptions(conn net.Conn) int {
	count := 0
	for _, subs := range b.channels {
		if subs[conn] {
			count++
		}
	}
	for _, subs := range b.patterns {
		if subs[conn] {
			count++
		}
	}
	return count
}

// encodeMessage returns a RESP2 array for a message push:
// *3\r\n$7\r\nmessage\r\n$<len>\r\n<channel>\r\n$<len>\r\n<payload>\r\n
func encodeMessage(channel, payload string) []byte {
	return encodeArray([]string{"message", channel, payload})
}

// encodePMessage returns a RESP2 array for a pattern message push:
// *4\r\n$8\r\npmessage\r\n$<len>\r\n<pattern>\r\n$<len>\r\n<channel>\r\n$<len>\r\n<payload>\r\n
func encodePMessage(pattern, channel, payload string) []byte {
	return encodeArray([]string{"pmessage", pattern, channel, payload})
}

// encodeArray encodes a slice of strings as a RESP2 bulk-string array.
func encodeArray(parts []string) []byte {
	buf := make([]byte, 0, 64)
	buf = append(buf, '*')
	buf = append(buf, itoa(len(parts))...)
	buf = append(buf, '\r', '\n')
	for _, p := range parts {
		buf = append(buf, '$')
		buf = append(buf, itoa(len(p))...)
		buf = append(buf, '\r', '\n')
		buf = append(buf, p...)
		buf = append(buf, '\r', '\n')
	}
	return buf
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	digits := make([]byte, 0, 10)
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
