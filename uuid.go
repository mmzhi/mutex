package mutex

import (
	"encoding/binary"
	"sync"
	"net"
	"crypto/rand"
	"time"
	"encoding/hex"
)

var (
	storageMutex  sync.Mutex
	storageOnce   sync.Once
	epochFunc     = unixTimeFunc
	clockSequence uint16
	lastTime      uint64
	hardwareAddr  [6]byte
)

const (
	DomainPerson = iota
	DomainGroup
	DomainOrg
)

const epochStart = 122192928000000000
const dash byte = '-'

func unixTimeFunc() uint64 {
	return epochStart + uint64(time.Now().UnixNano()/100)
}

func initClockSequence() {
	buf := make([]byte, 2)
	safeRandom(buf)
	clockSequence = binary.BigEndian.Uint16(buf)
}

func initHardwareAddr() {
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if len(iface.HardwareAddr) >= 6 {
				copy(hardwareAddr[:], iface.HardwareAddr)
				return
			}
		}
	}
	safeRandom(hardwareAddr[:])

	hardwareAddr[0] |= 0x01
}

func safeRandom(dest []byte) {
	if _, err := rand.Read(dest); err != nil {
		panic(err)
	}
}

func initStorage() {
	initClockSequence()
	initHardwareAddr()
}

func uuid() string{
	var u [16]byte

	storageOnce.Do(initStorage)

	storageMutex.Lock()
	defer storageMutex.Unlock()

	timeNow := epochFunc()
	if timeNow <= lastTime {
		clockSequence++
	}
	lastTime = timeNow


	binary.BigEndian.PutUint32(u[0:], uint32(timeNow))
	binary.BigEndian.PutUint16(u[4:], uint16(timeNow>>32))
	binary.BigEndian.PutUint16(u[6:], uint16(timeNow>>48))
	binary.BigEndian.PutUint16(u[8:], clockSequence)

	copy(u[10:], hardwareAddr[:])

	u[6] = (u[6] & 0x0f) | 16
	u[8] = (u[8] & 0xbf) | 0x80

	buf := make([]byte, 36)

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], u[10:])

	return string(buf)
}