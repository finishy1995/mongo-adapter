package protocol

import (
	"bytes"
	"encoding/binary"
	"io"

	"finishy1995/mongo-adapter/library/log"

	"go.mongodb.org/mongo-driver/bson"
)

type Conn interface {
	io.Reader
	io.Writer
}

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) OnMessage(conn Conn, buf []byte) bool {
	buffer := bytes.NewBuffer(buf)
	header := MsgHeader{}
	if len(buf) < int(header.MessageLength) {
		log.Errorf("Client sent less than %d bytes, len: %d", header.MessageLength, len(buf))
		return false
	}

	// 读取 header
	if err := binary.Read(buffer, binary.LittleEndian, &header); err != nil {
		log.Errorf("Error reading header: %v", err)
		return false
	}

	switch header.OpCode {
	case OP_QUERY: // OP_QUERY
		query := OpQuery{Header: header}
		if err := binary.Read(buffer, binary.LittleEndian, &query.Flags); err != nil {
			log.Errorf("Error reading flags:", err)
			return false
		}

		collectionName, err := readCString(buffer)
		if err != nil {
			log.Errorf("Error reading collection name:", err)
			return false
		}
		query.FullCollectionName = collectionName
		if err := binary.Read(buffer, binary.LittleEndian, &query.NumberToSkip); err != nil {
			log.Errorf("Error reading numberToSkip:", err)
			return false
		}
		if err := binary.Read(buffer, binary.LittleEndian, &query.NumberToReturn); err != nil {
			log.Errorf("Error reading numberToReturn:", err)
			return false
		}
		query.Query = buffer.Bytes()
		var cmd bson.D
		if err := bson.Unmarshal(query.Query, &cmd); err != nil {
			log.Errorf("Error unmarshalling query:", err)
			return false
		}

		log.Debugf("Received OP_QUERY requestID: %d, Collection: %s, Message: %+v", header.RequestID, query.FullCollectionName, cmd)

		if query.FullCollectionName == "admin.$cmd" {
			s.sendResponse(conn, header.RequestID, messageHandle(cmd))
		}
		break
	case OP_MSG:
		msg := OpMsg{Header: header}
		if err := binary.Read(buffer, binary.LittleEndian, &msg.Flags); err != nil {
			log.Errorf("Error reading flags:", err)
			return false
		}
		msg.Sections = make([]Section, 0)
		for buffer.Len() > 0 {
			kind, _ := buffer.ReadByte()
			if kind == 0 {
				// type 0, 主 BSON 文档
				docBytes, err := readBSONBytes(buffer)
				if err != nil {
					log.Errorf("Error reading type 0 BSON document:", err)
					return false
				}
				var doc bson.D
				bson.Unmarshal(docBytes, &doc)
				msg.Sections = append(msg.Sections, Section{Kind: 0, Body: doc})
			} else if kind == 1 {
				log.Errorf("Unsupported Kind == 1")
			} else {
				log.Errorf("Unsupported Kind == %d", kind)
			}
		}
		log.Debugf("Received OP_MSG requestID: %d, Message: %+v", header.RequestID, msg)
		s.sendResponse(conn, header.RequestID, messageHandle(msg.Sections[0].Body))
		break
	default:
		log.Errorf("Received unsupported OpCode: %d\n", header.OpCode)
		return false
	}

	return true
}

func (s *Server) sendResponse(conn Conn, requestID int32, responseDoc bson.M) {
	log.Debugf("sendResponse. requestID: %d, responseDoc: %+v", requestID, responseDoc)

	responseBytes, err := bson.Marshal(responseDoc)
	if err != nil {
		log.Errorf("Error marshaling response:", err)
		return
	}

	// Create response header
	header := MsgHeader{
		MessageLength: int32(36 + len(responseBytes)),
		RequestID:     requestID,
		ResponseTo:    requestID,
		OpCode:        OP_REPLY,
	}

	// Create reply message
	reply := OpReply{
		ResponseFlags:  0,
		CursorID:       0,
		StartingFrom:   0,
		NumberReturned: 1,
	}

	// Write header and reply
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, &header)
	binary.Write(buf, binary.LittleEndian, &reply.ResponseFlags)
	binary.Write(buf, binary.LittleEndian, &reply.CursorID)
	binary.Write(buf, binary.LittleEndian, &reply.StartingFrom)
	binary.Write(buf, binary.LittleEndian, &reply.NumberReturned)
	buf.Write(responseBytes)

	written, err := conn.Write(buf.Bytes())
	if err != nil || written != buf.Len() {
		log.Errorf("Error writing response: %v, written %d/%d bytes\n", err, written, buf.Len())
	}
}
