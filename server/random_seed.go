package server

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
)

var errRandomSeedGeneration = errors.New("failed to generate random seed")

var multiplayerRandomSeedSource = defaultMultiplayerRandomSeed

func defaultMultiplayerRandomSeed() (int64, error) {
	var bytes [8]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return 0, errRandomSeedGeneration
	}

	seed := int64(binary.LittleEndian.Uint64(bytes[:]) & 0x7fffffffffffffff)
	if seed == 0 {
		seed = 1
	}
	return seed, nil
}
