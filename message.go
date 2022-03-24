package earth

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"github.com/victor-leee/earth/github.com/victor-leee/earth"
	"google.golang.org/protobuf/proto"
	"io"
)

// Message holds the data transferred through unix domain sock and sends them to other side-cars
type Message struct {
	HeaderLenBytes []byte
	Header         *earth.Header
	RawHeader      []byte
	Body           []byte
}

func FromProtoMessage(msg proto.Message, header *earth.Header) *Message {
	b, _ := proto.Marshal(msg)
	return FromBody(b, header)
}

// FromBody builds a Message directly with information of the body bytes(build header&headerLength)
func FromBody(body []byte, customHeader *earth.Header) *Message {
	if customHeader == nil {
		customHeader = &earth.Header{}
	}
	header := &earth.Header{
		BodySize:            uint64(len(body)),
		MessageType:         customHeader.MessageType,
		SenderServiceName:   customHeader.SenderServiceName,
		ReceiverServiceName: customHeader.ReceiverServiceName,
		TraceId:             customHeader.TraceId,
		Extra:               customHeader.Extra,
	}
	headerBytes, _ := proto.Marshal(header)
	headerLenBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(headerLenBytes, uint64(len(headerBytes)))

	return &Message{
		HeaderLenBytes: headerLenBytes,
		Header:         header,
		RawHeader:      headerBytes,
		Body:           body,
	}
}

// FromReader builds a Message by calling readFunc on reader for multiple times
func FromReader(reader io.Reader, readFunc func(reader io.Reader, size uint64) ([]byte, error)) (*Message, error) {
	// first block read the first 8 bytes, which is the length of the header
	headerLenBytes, err := readFunc(reader, 8)
	if err != nil {
		logrus.Errorf("[FromReader] read header length bytes failed: %v", err)
		return nil, err
	}
	headerLen := binary.LittleEndian.Uint64(headerLenBytes)
	// then block read the header with length of headerLen
	headerBytes, err := readFunc(reader, headerLen)
	if err != nil {
		logrus.Errorf("[FromReader] reader header bytes failed: %v", err)
		return nil, err
	}
	header := &earth.Header{}
	if err = proto.Unmarshal(headerBytes, header); err != nil {
		logrus.Errorf("[FromReader] unmarshal bytes to struct Header failed: %v", err)
		return nil, err
	}
	// eventually read the body bytes
	body, err := readFunc(reader, header.BodySize)
	if err != nil {
		logrus.Errorf("[FromReader] reader body failed: %v", err)
		return nil, err
	}

	return &Message{
		HeaderLenBytes: headerLenBytes,
		Header:         header,
		RawHeader:      headerBytes,
		Body:           body,
	}, nil
}

func (m *Message) Write(writer io.Writer) (int, error) {
	var (
		nWrite int
		nInc   int
		err    error
	)

	nInc, err = writer.Write(m.HeaderLenBytes)
	if err != nil {
		logrus.Errorf("[Message.Write] write header length bytes failed: %v", err)
		return 0, err
	}
	nWrite += nInc

	nInc, err = writer.Write(m.RawHeader)
	if err != nil {
		logrus.Errorf("[Message.Write] write header bytes failed: %v", err)
		return 0, err
	}
	nWrite += nInc

	nInc, err = writer.Write(m.Body)
	if err != nil {
		logrus.Errorf("[Message.Write] write body bytes failed: %v", err)
		return 0, err
	}
	nWrite += nInc

	return nWrite, nil
}
