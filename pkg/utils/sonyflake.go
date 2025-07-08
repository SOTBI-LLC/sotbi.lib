package utils

import (
	"errors"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

// SonyflakeConfig holds configuration for Sonyflake.
type SonyflakeConfig struct {
	MachineID   uint8 // 8 bits (0-255)
	CustomEpoch int64 // milliseconds since Unix epoch
}

// Sonyflake is a goroutine-safe Sonyflake ID generator using atomic state.
type Sonyflake struct {
	state  uint64 // high 48 bits: time, low 16 bits: sequence
	config SonyflakeConfig
}

const (
	sonyflakeMachineIDBits = 8
	sonyflakeSequenceBits  = 8
	sonyflakeTimeBits      = 48
	sonyflakeMaxSequence   = 1<<sonyflakeSequenceBits - 1
	sonyflakeTimeMask      = (uint64(1)<<sonyflakeTimeBits - 1) << sonyflakeSequenceBits
	sonyflakeSeqMask       = uint64(sonyflakeMaxSequence)
)

// NewSonyflake creates a new Sonyflake instance.
func NewSonyflake(cfg SonyflakeConfig) *Sonyflake {
	if cfg.CustomEpoch == 0 {
		cfg.CustomEpoch = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	}

	return &Sonyflake{
		config: cfg,
	}
}

// NextID generates a new unique Sonyflake ID (goroutine-safe, atomic).
func (s *Sonyflake) NextID() (uint64, error) {
	machineID := uint64(s.config.MachineID)

	for {
		now := time.Now().UnixMilli() - s.config.CustomEpoch
		if now < 0 {
			return 0, errors.New("time is before custom epoch")
		}

		if now >= (1 << sonyflakeTimeBits) {
			return 0, errors.New("time overflow")
		}

		curr := atomic.LoadUint64(&s.state)
		lastTime := int64((curr & sonyflakeTimeMask) >> sonyflakeSequenceBits)
		seq := uint16(curr & sonyflakeSeqMask)

		if now == lastTime { //nolint:gocritic
			if seq == sonyflakeMaxSequence {
				// Sequence overflow, wait for next millisecond
				for now <= lastTime {
					now = time.Now().UnixMilli() - s.config.CustomEpoch
				}

				continue
			}

			newSeq := seq + 1
			newState := (uint64(now) << sonyflakeSequenceBits) | uint64(newSeq)

			if atomic.CompareAndSwapUint64(&s.state, curr, newState) {
				return (uint64(now) << (sonyflakeMachineIDBits + sonyflakeSequenceBits)) |
					(machineID << sonyflakeSequenceBits) |
					uint64(newSeq), nil
			}
		} else if now > lastTime {
			// New millisecond, reset sequence to random start
			newSeq := uint16(rand.IntN(2)) //nolint:gosec
			newState := (uint64(now) << sonyflakeSequenceBits) | uint64(newSeq)

			if atomic.CompareAndSwapUint64(&s.state, curr, newState) {
				return (uint64(now) << (sonyflakeMachineIDBits + sonyflakeSequenceBits)) |
					(machineID << sonyflakeSequenceBits) |
					uint64(newSeq), nil
			}
		} else {
			// Clock moved backwards, rare, spin
			continue
		}
	}
}
