package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser parses RESP protocol messages
type Parser struct {
	reader *bufio.Reader
}

// NewParser creates a new RESP parser
func NewParser(reader io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(reader),
	}
}

// Parse reads and parses the next RESP value
func (parser *Parser) Parse() (Value, error) {
	typeByte, err := parser.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch Type(typeByte) {
	case SimpleString:
		return parser.parseSimpleString()
	case Error:
		return parser.parseError()
	case Integer:
		return parser.parseInteger()
	case BulkString:
		return parser.parseBulkString()
	case Array:
		return parser.parseArray()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %c", typeByte)
	}
}

func (parser *Parser) readLine() (string, error) {
	line, err := parser.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	// Remove \r\n
	return strings.TrimSuffix(line, "\r\n"), nil
}

func (parser *Parser) parseSimpleString() (Value, error) {
	line, err := parser.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: SimpleString, Str: line}, nil
}

func (parser *Parser) parseError() (Value, error) {
	line, err := parser.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: Error, Str: line}, nil
}

func (parser *Parser) parseInteger() (Value, error) {
	line, err := parser.readLine()
	if err != nil {
		return Value{}, err
	}

	intValue, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid integer: %s", line)
	}

	return Value{Type: Integer, Integer: intValue}, nil
}

func (parser *Parser) parseBulkString() (Value, error) {
	line, err := parser.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk string length: %s", line)
	}

	if length == -1 {
		// Null bulk string - use special marker
		return Value{Type: BulkString, Str: "\x00NULL"}, nil
	}

	if length < 0 {
		return Value{}, fmt.Errorf("invalid bulk string length: %d", length)
	}

	// Read the bulk string data
	data := make([]byte, length+2) // +2 for \r\n
	_, err = io.ReadFull(parser.reader, data)
	if err != nil {
		return Value{}, err
	}

	// Remove \r\n
	return Value{Type: BulkString, Str: string(data[:length])}, nil
}

func (parser *Parser) parseArray() (Value, error) {
	line, err := parser.readLine()
	if err != nil {
		return Value{}, err
	}

	count, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid array count: %s", line)
	}

	if count == -1 {
		// Null array
		return Value{Type: Array, Array: []Value{}}, nil
	}

	if count < 0 {
		return Value{}, fmt.Errorf("invalid array count: %d", count)
	}

	array := make([]Value, count)
	for index := 0; index < count; index++ {
		value, err := parser.Parse()
		if err != nil {
			return Value{}, err
		}
		array[index] = value
	}

	return Value{Type: Array, Array: array}, nil
}
