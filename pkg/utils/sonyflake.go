package utils

import (
	"errors"
	"math/rand/v2"
	"sync/atomic"
	"time"
)

// SonyflakeConfig holds configuration for Sonyflake.
type SonyflakeConfig struct {
	MachineID   uint8 // 6 bits (0-63) - reduced for JS compatibility
	CustomEpoch int64 // milliseconds since Unix epoch
}

// Sonyflake is a goroutine-safe Sonyflake ID generator using atomic state.
// JavaScript-compatible: generates 53-bit IDs that fit in Number.MAX_SAFE_INTEGER.
type Sonyflake struct {
	state  uint64 // high 41 bits: time, low 6 bits: sequence
	config SonyflakeConfig
}

const (
	// JavaScript-compatible bit allocation (53 bits total).
	sonyflakeMachineIDBits = 6                             // 6 bits: 0-63 machines
	sonyflakeSequenceBits  = 6                             // 6 bits: 0-63 per millisecond
	sonyflakeTimeBits      = 41                            // 41 bits: ~69 years from epoch
	sonyflakeMaxSequence   = 1<<sonyflakeSequenceBits - 1  // 63
	sonyflakeMaxMachineID  = 1<<sonyflakeMachineIDBits - 1 // 63
	sonyflakeMaxTime       = 1<<sonyflakeTimeBits - 1      // ~69 years
	sonyflakeTimeMask      = (uint64(1)<<sonyflakeTimeBits - 1) << sonyflakeSequenceBits
	sonyflakeSeqMask       = uint64(sonyflakeMaxSequence)
)

// NewSonyflake creates a new Sonyflake instance.
func NewSonyflake(cfg SonyflakeConfig) (*Sonyflake, error) {
	if cfg.MachineID > sonyflakeMaxMachineID {
		return nil, errors.New("machine ID exceeds maximum value (0-63)")
	}

	if cfg.CustomEpoch == 0 {
		cfg.CustomEpoch = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	}

	return &Sonyflake{
		config: cfg,
	}, nil
}

// NextID generates a new unique Sonyflake ID (goroutine-safe, atomic).
// Returns a 53-bit ID that's safe to use in JavaScript.
func (s *Sonyflake) NextID() (uint64, error) {
	machineID := uint64(s.config.MachineID)

	for {
		now := time.Now().UnixMilli() - s.config.CustomEpoch
		if now < 0 {
			return 0, errors.New("time is before custom epoch")
		}

		if now > sonyflakeMaxTime {
			return 0, errors.New("time overflow: exceeds 41-bit limit")
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
			newSeq := uint16(rand.IntN(sonyflakeMaxSequence + 1)) //nolint:gosec
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

// MaxSafeInteger returns the maximum safe integer for JavaScript (2^53 - 1).
func (s *Sonyflake) MaxSafeInteger() uint64 {
	return 9007199254740991 // 2^53 - 1
}

// IsJavaScriptSafe checks if the given ID is safe to use in JavaScript.
func (s *Sonyflake) IsJavaScriptSafe(id uint64) bool {
	return id <= s.MaxSafeInteger()
}
