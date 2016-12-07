/*
MIT License

Copyright 2016 Comcast Cable Communications Management, LLC

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Package scte35 is for handling scte35 splice signals
package scte35

import (
	"github.com/Philoinc/gots"
	"github.com/Philoinc/gots/psi"
)

type SpliceCommandType uint16

const (
	// SpliceNull is a Null Splice command type
	SpliceNull SpliceCommandType = 0x00
	// SpliceSchedule is a splice schedule command type
	SpliceSchedule = 0x04
	// SpliceInsert is a splice insert command type
	SpliceInsert = 0x05
	// TimeSignal is a splice signal command type
	TimeSignal = 0x06
	// BandwidthReservation is a command type that represents a reservation of bandwidth
	BandwidthReservation = 0x07
	// PrivateCommand is a command type that represents private command data
	PrivateCommand = 0xFF
)

// SegDescType is the Segmentation Descriptor Type - not really needed for processing according
// to method below, but included here for backwards compatibility/porting
type SegDescType uint8

const (
	SegDescNotIndicated                  SegDescType = 0x00
	SegDescContentIdentification                     = 0x01
	SegDescProgramStart                              = 0x10
	SegDescProgramEnd                                = 0x11
	SegDescProgramEarlyTermination                   = 0x12
	SegDescProgramBreakaway                          = 0x13
	SegDescProgramResumption                         = 0x14
	SegDescProgramRunoverPlanned                     = 0x15
	SegDescProgramRunoverUnplanned                   = 0x16
	SegDescProgramOverlapStart                       = 0x17
	SegDescProgramBlackoutOverride                   = 0x18
	SegDescChapterStart                              = 0x20
	SegDescChapterEnd                                = 0x21
	SegDescProviderAdvertisementStart                = 0x30
	SegDescProviderAdvertisementEnd                  = 0x31
	SegDescDistributorAdvertisementStart             = 0x32
	SegDescDistributorAdvertisementEnd               = 0x33
	SegDescProviderPOStart                           = 0x34
	SegDescProviderPOEnd                             = 0x35
	SegDescDistributorPOStart                        = 0x36
	SegDescDistributorPOEnd                          = 0x37
	SegDescUnscheduledEventStart                     = 0x40
	SegDescUnscheduledEventEnd                       = 0x41
	SegDescNetworkStart                              = 0x50
	SegDescNetworkEnd                                = 0x51
)

// SegUPIDType is the Segmentation UPID Types - Only type that really needs to be checked is
// SegUPIDURN for CSP
type SegUPIDType uint8

const (
	SegUPIDNotUsed     SegUPIDType = 0x00
	SegUPIDUserDefined             = 0x01
	SegUPIDISCI                    = 0x02
	SegUPIDAdID                    = 0x03
	SegUPIDUMID                    = 0x04
	SegUPIDISAN                    = 0x05
	SegUPIDVISAN                   = 0x06
	SegUPIDTID                     = 0x07
	SegUPIDTI                      = 0x08
	SegUPIDADI                     = 0x09
	SegUPIDEIDR                    = 0x0a
	SegUPIDATSCID                  = 0x0b
	SegUPIDMPU                     = 0x0c
	SegUPIDMID                     = 0x0d
	SegUPADSINFO                   = 0x0e
	SegUPIDURN                     = 0x0f
)

// SCTE35 represent operations available on a SCTE35 message.
type SCTE35 interface {
	// HasPTS returns true if there is a pts time
	HasPTS() bool
	// PTS returns the PTS time of the signal if it exists
	PTS() gots.PTS
	// Command returns the signal's splice command
	Command() SpliceCommandType
	// CommandInfo returns an object describing fields of the signal's splice
	// command structure
	CommandInfo() SpliceCommand
	// Descriptors returns a slice of the signals SegmentationDescriptors sorted
	// by descriptor weight (least important signals first)
	Descriptors() []SegmentationDescriptor
	// Data returns the raw data bytes of the scte signal
	Data() []byte
}

type SpliceCommand interface {
	CommandType() SpliceCommandType
}

type PTSCommand interface {
	SpliceCommand
	HasPTS() bool
	PTS() gots.PTS
}

type TimeSignalCommand interface {
	PTSCommand
}

type SpliceInsertCommand interface {
	SpliceCommand
	EventCancelIndicator() bool
	OutOfNetworkIndicator() bool
	EventID() uint32
	HasPTS() bool
	PTS() gots.PTS
	HasDuration() bool
	Duration() gots.PTS
	AutoReturn() bool
}

// SegmentationDescriptor describes the segmentation descriptor interface.
type SegmentationDescriptor interface {
	// SCTE35 returns the SCTE35 signal this segmentation descriptor was found in.
	SCTE35() SCTE35
	// EventID returns the event id
	EventID() uint32
	// EventCancelIndicator returns whether the
	// segmentation_event_cancel_indicator bit is set
	EventCancelIndicator() bool
	// TypeID returns the segmentation type for descriptor
	TypeID() SegDescType
	// IsOut returns true if a signal is an out
	IsOut() bool
	// IsIn returns true if a signal is an in
	IsIn() bool
	// HasDuration returns true if there is a duration associated with the descriptor
	HasDuration() bool
	// Duration returns the duration of the descriptor
	Duration() gots.PTS
	// UPIDType returns the type of the upid
	UPIDType() SegUPIDType
	// UPID returns the upid of the descriptor
	UPID() []byte
	// CanClose returns true if this descriptor can close the passed in descriptor
	CanClose(out SegmentationDescriptor) bool
	// Equal returns true/false if segmentation descriptor is functionally
	// equal (i.e. a duplicate)
	Equal(sd SegmentationDescriptor) bool
}

// State maintains current state for all signals and descriptors.  The intended
// usage is to call ParseSCTE35() on raw data to create a signal, and then call
// ProcessSignal with that signal.  This returns the list of descriptors closed
// by that signal. If signals have a duration and need to be closed implicitly
// after some timer has passed, then Close() can be used for that.  Some
// example code is below.
// s := scte35.NewState()
// scte,_ := scte.ParseSCTE35(bytes)
// for _,d := range(scte.Descriptors()) {
//   closed = s.ProcessDescriptor(d)
//   ...handle closed signals appropriately here
//   if d.HasDuration() {
//     time.AfterFunc(d.Duration() + someFudgeDelta,
//                    func() { closed = s.Close(d) })
//   }
// }
type State interface {
	// Open returns a list of open signals
	Open() []SegmentationDescriptor
	// Process takes a scte35 descriptor and returns a list of descriptors closed by it
	ProcessDescriptor(desc SegmentationDescriptor) ([]SegmentationDescriptor, error)
	// Close acts like Process and acts as if an appropriate close has been
	// received for this given descriptor.
	Close(desc SegmentationDescriptor) ([]SegmentationDescriptor, error)
}

// SCTE done func is the same as the PMT because they're both psi
func SCTE35AccumulatorDoneFunc(b []byte) (bool, error) {
	return psi.PmtAccumulatorDoneFunc(b)
}
