package rdb

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/codecrafters-redis-go/internal/storage"
)

const (
	// RDB magic string
	rdbMagic = "REDIS"

	// Op codes
	opEOF          = 0xFF
	opSelectDB     = 0xFE
	opExpireTime   = 0xFD
	opExpireTimeMs = 0xFC
	opResizeDB     = 0xFB
	opAux          = 0xFA

	// String encoding types
	stringTypeLen  = 0x00 // Length prefixed string
	stringTypeInt8 = 0xC0 // 8 bit integer
	stringTypeInt16 = 0xC1 // 16 bit integer
	stringTypeInt32 = 0xC2 // 32 bit integer
	stringTypeLZF   = 0xC3 // LZF compressed string

	// Value types
	valueTypeString = 0
)

// Loader loads data from RDB files
type Loader struct {
	reader  io.Reader
	storage *storage.Storage
}

// LoadFile loads an RDB file into storage
func LoadFile(dir, filename string, store *storage.Storage) error {
	path := filepath.Join(dir, filename)

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// No RDB file, that's ok
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open RDB file: %w", err)
	}
	defer file.Close()

	loader := &Loader{
		reader:  file,
		storage: store,
	}

	return loader.load()
}

func (loader *Loader) load() error {
	// Read and verify magic string
	magic := make([]byte, 5)
	if _, err := io.ReadFull(loader.reader, magic); err != nil {
		return fmt.Errorf("failed to read magic: %w", err)
	}

	if string(magic) != rdbMagic {
		return fmt.Errorf("invalid RDB file: wrong magic string")
	}

	// Read version (4 bytes)
	version := make([]byte, 4)
	if _, err := io.ReadFull(loader.reader, version); err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}

	// Process the RDB file
	for {
		// Read op code
		opCode, err := loader.readByte()
		if err != nil {
			return fmt.Errorf("failed to read op code: %w", err)
		}

		switch opCode {
		case opEOF:
			// End of file
			return nil

		case opSelectDB:
			// Select database (we ignore this for now)
			if _, err := loader.readLength(); err != nil {
				return err
			}

		case opResizeDB:
			// Database size hint (we ignore this)
			if _, err := loader.readLength(); err != nil {
				return err
			}
			if _, err := loader.readLength(); err != nil {
				return err
			}

		case opAux:
			// Auxiliary field (we ignore this)
			if _, err := loader.readString(); err != nil {
				return err
			}
			if _, err := loader.readString(); err != nil {
				return err
			}

		case opExpireTimeMs:
			// Millisecond precision expiry
			expiryMs, err := loader.readUint64()
			if err != nil {
				return err
			}

			// Read the key-value pair
			if err := loader.readKeyValue(expiryMs); err != nil {
				return err
			}

		case opExpireTime:
			// Second precision expiry
			expirySec, err := loader.readUint32()
			if err != nil {
				return err
			}

			// Convert to milliseconds
			expiryMs := uint64(expirySec) * 1000

			// Read the key-value pair
			if err := loader.readKeyValue(expiryMs); err != nil {
				return err
			}

		default:
			// This is a value type
			if err := loader.readValue(opCode, 0); err != nil {
				return err
			}
		}
	}
}

func (loader *Loader) readKeyValue(expiryMs uint64) error {
	// Read value type
	valueType, err := loader.readByte()
	if err != nil {
		return err
	}

	return loader.readValue(valueType, expiryMs)
}

func (loader *Loader) readValue(valueType byte, expiryMs uint64) error {
	// For now, we only support string values
	if valueType != valueTypeString {
		return fmt.Errorf("unsupported value type: %d", valueType)
	}

	// Read key
	key, err := loader.readString()
	if err != nil {
		return fmt.Errorf("failed to read key: %w", err)
	}

	// Read value
	value, err := loader.readString()
	if err != nil {
		return fmt.Errorf("failed to read value: %w", err)
	}

	// Calculate expiration
	var expiration *time.Time
	if expiryMs > 0 {
		expiryTime := time.UnixMilli(int64(expiryMs))
		expiration = &expiryTime
	}

	// Store in our storage
	loader.storage.Set(key, value, expiration)

	return nil
}

func (loader *Loader) readByte() (byte, error) {
	buf := make([]byte, 1)
	if _, err := io.ReadFull(loader.reader, buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (loader *Loader) readUint32() (uint32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(loader.reader, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

func (loader *Loader) readUint64() (uint64, error) {
	buf := make([]byte, 8)
	if _, err := io.ReadFull(loader.reader, buf); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf), nil
}

func (loader *Loader) readLength() (uint64, error) {
	firstByte, err := loader.readByte()
	if err != nil {
		return 0, err
	}

	// Check encoding type
	encType := (firstByte & 0xC0) >> 6

	switch encType {
	case 0:
		// Next 6 bits represent the length
		return uint64(firstByte & 0x3F), nil

	case 1:
		// Read one more byte, combined 14 bits represent the length
		nextByte, err := loader.readByte()
		if err != nil {
			return 0, err
		}
		return uint64((firstByte&0x3F)<<8) | uint64(nextByte), nil

	case 2:
		// Read 4 more bytes
		buf := make([]byte, 4)
		if _, err := io.ReadFull(loader.reader, buf); err != nil {
			return 0, err
		}
		return uint64(binary.BigEndian.Uint32(buf)), nil

	case 3:
		// Special encoding - return the byte as-is
		return uint64(firstByte), nil

	default:
		return 0, fmt.Errorf("unexpected encoding type")
	}
}

func (loader *Loader) readString() (string, error) {
	length, err := loader.readLength()
	if err != nil {
		return "", err
	}

	// Check if it's a special encoding (when encType was 3)
	if length >= 0xC0 {
		// Special encoding (integers)
		switch byte(length) {
		case stringTypeInt8:
			b, err := loader.readByte()
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", int8(b)), nil

		case stringTypeInt16:
			buf := make([]byte, 2)
			if _, err := io.ReadFull(loader.reader, buf); err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", int16(binary.LittleEndian.Uint16(buf))), nil

		case stringTypeInt32:
			buf := make([]byte, 4)
			if _, err := io.ReadFull(loader.reader, buf); err != nil {
				return "", err
			}
			return fmt.Sprintf("%d", int32(binary.LittleEndian.Uint32(buf))), nil

		default:
			return "", fmt.Errorf("unsupported string encoding: %d", length)
		}
	}

	// Regular string
	buf := make([]byte, length)
	if _, err := io.ReadFull(loader.reader, buf); err != nil {
		return "", err
	}

	return string(buf), nil
}
