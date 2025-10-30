// Copyright (c) The GoTEE authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package main

import (
	_ "unsafe"

	"github.com/usbarmory/GoTEE-example/mem"
)

//go:linkname ramStart runtime.ramStart
var ramStart uint32 = mem.AppletVirtualStart

//go:linkname ramSize runtime.ramSize
var ramSize uint32 = mem.AppletSize

//go:linkname ramStackOffset runtime.ramStackOffset
var ramStackOffset uint32 = 0x100
