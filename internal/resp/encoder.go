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
func NewEncoder(writer io.Writer) *Encoder {
	return &Encoder{writer: writer}
}

// Encode writes a RESP value to the writer
func (encoder *Encoder) Encode(value Value) error {
	switch value.Type {
	case SimpleString:
		return encoder.encodeSimpleString(value.Str)
	case Error:
		return encoder.encodeError(value.Str)
	case Integer:
		return encoder.encodeInteger(value.Integer)
	case BulkString:
		return encoder.encodeBulkString(value.Str)
	case Array:
		return encoder.encodeArray(value.Array)
	default:
		return fmt.Errorf("unknown RESP type: %c", value.Type)
	}
}

func (encoder *Encoder) write(data string) error {
	_, err := encoder.writer.Write([]byte(data))
	return err
}

func (encoder *Encoder) encodeSimpleString(str string) error {
	return encoder.write(fmt.Sprintf("+%s\r\n", str))
}

func (encoder *Encoder) encodeError(str string) error {
	return encoder.write(fmt.Sprintf("-%s\r\n", str))
}

func (encoder *Encoder) encodeInteger(intValue int) error {
	return encoder.write(fmt.Sprintf(":%d\r\n", intValue))
}

func (encoder *Encoder) encodeBulkString(str string) error {
	// Check for null bulk string (special marker)
	if str == "\x00NULL" {
		return encoder.write("$-1\r\n")
	}
	return encoder.write(fmt.Sprintf("$%d\r\n%s\r\n", len(str), str))
}

func (encoder *Encoder) encodeArray(array []Value) error {
	if err := encoder.write(fmt.Sprintf("*%d\r\n", len(array))); err != nil {
		return err
	}

	for _, value := range array {
		if err := encoder.Encode(value); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions for common responses

// SimpleString creates a simple string value
func SimpleStringValue(str string) Value {
	return Value{Type: SimpleString, Str: str}
}

// Error creates an error value
func ErrorValue(str string) Value {
	return Value{Type: Error, Str: str}
}

// Integer creates an integer value
func IntegerValue(intValue int) Value {
	return Value{Type: Integer, Integer: intValue}
}

// BulkString creates a bulk string value
func BulkStringValue(str string) Value {
	return Value{Type: BulkString, Str: str}
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
