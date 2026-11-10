// Copyright 2015-present, The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package packet exercises parameters named after their package, which the
// generated cgo wrappers use to qualify calls (issue #381).
package packet

import (
	"bytes"
	"fmt"
)

type Packet interface {
	Marshal(w *bytes.Buffer)
}

type Ping struct {
	ID int
}

func NewPing(id int) *Ping {
	return &Ping{ID: id}
}

func (p *Ping) Marshal(w *bytes.Buffer) {
	fmt.Fprintf(w, "ping:%d", p.ID)
}

// GoPacketToBytes is the function from issue #381: an interface-typed
// parameter named after the package.
func GoPacketToBytes(shieldID int, packet Packet) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	fmt.Fprintf(buf, "%d/", shieldID)
	packet.Marshal(buf)
	return buf.Bytes(), nil
}

// PingID takes a struct-typed parameter named after the package.
func PingID(packet *Ping) int {
	return packet.ID
}

// Double takes a basic-typed parameter named after the package.
func Double(packet int) int {
	return 2 * packet
}

// Add takes a parameter named after the package, on a method.
func (p *Ping) Add(packet int) int {
	return p.ID + packet
}

// Apply takes a callback parameter named after the package.
func Apply(n int, packet func(i int) int) int {
	return packet(n)
}
