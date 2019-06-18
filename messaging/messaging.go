package messaging

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"

	"github.com/pkg/errors"
)

type Message struct {
	URL    string `json:"jnlp,omitempty"`
	Status string `json:"status,omitempty"`
}

// GetMessage gets a message from a connected browser extention
func GetMessage(reader io.Reader) (message *Message, err error) {
	defer func() {
		if err != nil {
			err = errors.Wrap(err, "error while receiving message")
		}
	}()
	var dataLen int32
	if err = binary.Read(reader, binary.LittleEndian, &dataLen); err != nil {
		return
	}
	data := make([]byte, dataLen)
	if _, err = reader.Read(data); err != nil {
		return
	}
	log.Printf("got message %v\n", string(data))
	var msg Message
	if err = json.Unmarshal(data, &msg); err != nil {
		return
	}
	return &msg, nil
}

// SendMessage sends a to a connected browser extention
func SendMessage(writer io.Writer, message string) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "error while sending message '%s'", message)
		}
	}()
	// messageLen := int32(len(message))
	buffer := new(bytes.Buffer)
	if err = binary.Write(buffer, binary.LittleEndian, int32(len(message))); err != nil {
		return
	}
	if _, err = buffer.WriteString(message); err != nil {
		return
	}
	if _, err = buffer.WriteTo(writer); err != nil {
		return
	}
	return nil
}
