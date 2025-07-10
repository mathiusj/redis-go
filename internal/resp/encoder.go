package resp

import (
	"fmt"
	"io"
)

// Encoder encodes values to RESP format
type Encoder struct {
	writer io.Writer
}

// NewEncoder creates a new RESP encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{writer: w}
}

// Encode writes a RESP value to the writer
func (e *Encoder) Encode(v Value) error {
	switch v.Type {
	case SimpleString:
		return e.encodeSimpleString(v.Str)
	case Error:
		return e.encodeError(v.Str)
	case Integer:
		return e.encodeInteger(v.Integer)
	case BulkString:
		return e.encodeBulkString(v.Str)
	case Array:
		return e.encodeArray(v.Array)
	default:
		return fmt.Errorf("unknown RESP type: %c", v.Type)
	}
}

func (e *Encoder) write(data string) error {
	_, err := e.writer.Write([]byte(data))
	return err
}

func (e *Encoder) encodeSimpleString(s string) error {
	return e.write(fmt.Sprintf("+%s\r\n", s))
}

func (e *Encoder) encodeError(s string) error {
	return e.write(fmt.Sprintf("-%s\r\n", s))
}

func (e *Encoder) encodeInteger(i int) error {
	return e.write(fmt.Sprintf(":%d\r\n", i))
}

func (e *Encoder) encodeBulkString(s string) error {
	// Check for null bulk string (special marker)
	if s == "\x00NULL" {
		return e.write("$-1\r\n")
	}
	return e.write(fmt.Sprintf("$%d\r\n%s\r\n", len(s), s))
}

func (e *Encoder) encodeArray(array []Value) error {
	if err := e.write(fmt.Sprintf("*%d\r\n", len(array))); err != nil {
		return err
	}

	for _, v := range array {
		if err := e.Encode(v); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions for common responses

// SimpleString creates a simple string value
func SimpleStringValue(s string) Value {
	return Value{Type: SimpleString, Str: s}
}

// Error creates an error value
func ErrorValue(s string) Value {
	return Value{Type: Error, Str: s}
}

// Integer creates an integer value
func IntegerValue(i int) Value {
	return Value{Type: Integer, Integer: i}
}

// BulkString creates a bulk string value
func BulkStringValue(s string) Value {
	return Value{Type: BulkString, Str: s}
}

// Array creates an array value
func ArrayValue(values ...Value) Value {
	return Value{Type: Array, Array: values}
}

// NullBulkString creates a null bulk string value
func NullBulkString() Value {
	// Use a special marker to indicate null bulk string
	return Value{Type: BulkString, Str: "\x00NULL"}
}

// OK returns a standard OK simple string
func OK() Value {
	return SimpleStringValue("OK")
}

// Pong returns a standard PONG simple string
func Pong() Value {
	return SimpleStringValue("PONG")
}
