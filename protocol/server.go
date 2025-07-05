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
	if len(buf) < 16 {
		log.Errorf("Client sent short packet, len: %d", len(buf))
		return false
	}
	header := MsgHeader{}
	if err := binary.Read(buffer, binary.LittleEndian, &header); err != nil {
		log.Errorf("Error reading header: %v", err)
		return false
	}
	if len(buf) < int(header.MessageLength) {
		log.Errorf("Client sent less than header.MessageLength: %d", header.MessageLength)
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
			s.sendResponse(conn, header.RequestID, header.OpCode, messageHandle(cmd))
		}
		break
	case OP_MSG:
		msg := OpMsg{Header: header}
		msgLength := int(header.MessageLength)
		if len(buf) < msgLength {
			log.Errorf("buf too short for OP_MSG")
			return false
		}
		body := buf[16:msgLength] // 只处理当前消息体
		offset := 0

		if len(body) < 4 {
			log.Errorf("body too short for flags")
			return false
		}
		msg.Flags = binary.LittleEndian.Uint32(body[offset : offset+4])
		offset += 4

		msg.Sections = make([]Section, 0)
		for offset < len(body) {
			kind := body[offset]
			offset++
			if kind == 0 {
				if offset+4 > len(body) {
					log.Errorf("Not enough bytes for BSON length")
					return false
				}
				docLen := int(binary.LittleEndian.Uint32(body[offset : offset+4]))
				if offset+docLen > len(body) {
					log.Errorf("BSON out of bounds")
					return false
				}
				docBytes := body[offset : offset+docLen]
				var doc bson.D
				if err := bson.Unmarshal(docBytes, &doc); err != nil {
					log.Errorf("Error decoding BSON: %v", err)
					return false
				}
				msg.Sections = append(msg.Sections, Section{Kind: 0, Body: doc})
				offset += docLen
			} else if kind == 1 {
				log.Errorf("Unsupported Kind == 1")
				return false
			} else {
				log.Errorf("Unsupported Kind == %d", kind)
				return false
			}
		}
		log.Debugf("Received OP_MSG requestID: %d, Message: %+v", header.RequestID, msg)
		s.sendResponse(conn, header.RequestID, header.OpCode, messageHandle(msg.Sections[0].Body))
		break
	default:
		log.Errorf("Received unsupported OpCode: %d\n", header.OpCode)
		return false
	}

	return true
}
func (s *Server) sendResponse(conn Conn, requestID int32, requestOpCode int32, responseDoc bson.M) {
	log.Debugf("sendResponse. requestID: %d, responseDoc: %+v", requestID, responseDoc)

	responseBytes, err := bson.Marshal(responseDoc)
	if err != nil {
		log.Errorf("Error marshaling response: %v", err)
		return
	}

	// MongoDB 3.6- 之前的版本，返回的是 OP_REPLY
	if requestOpCode == OP_QUERY {
		// --- OP_REPLY 格式 ---
		header := MsgHeader{
			MessageLength: int32(36 + len(responseBytes)), // header(16) + reply(20) + bson
			RequestID:     requestID,
			ResponseTo:    requestID,
			OpCode:        OP_REPLY,
		}
		reply := OpReply{
			ResponseFlags:  0,
			CursorID:       0,
			StartingFrom:   0,
			NumberReturned: 1,
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, &header)
		binary.Write(buf, binary.LittleEndian, &reply.ResponseFlags)
		binary.Write(buf, binary.LittleEndian, &reply.CursorID)
		binary.Write(buf, binary.LittleEndian, &reply.StartingFrom)
		binary.Write(buf, binary.LittleEndian, &reply.NumberReturned)
		buf.Write(responseBytes)

		written, err := conn.Write(buf.Bytes())
		if err != nil || written != buf.Len() {
			log.Errorf("Error writing OP_REPLY response: %v, written %d/%d bytes\n", err, written, buf.Len())
		}
	} else if requestOpCode == OP_MSG {
		// --- OP_MSG 格式 ---
		// OP_MSG header(16) + flags(4) + section0_kind(1) + bson
		flags := int32(0)
		section0Kind := byte(0)

		messageLength := int32(21 + len(responseBytes))
		header := MsgHeader{
			MessageLength: messageLength,
			RequestID:     requestID,
			ResponseTo:    requestID,
			OpCode:        OP_MSG,
		}

		buf := new(bytes.Buffer)
		binary.Write(buf, binary.LittleEndian, &header)
		binary.Write(buf, binary.LittleEndian, flags) // Flags (4 bytes)
		buf.WriteByte(section0Kind)                   // Section 0 Kind (1 byte)
		buf.Write(responseBytes)                      // Section 0 Body (bson)

		written, err := conn.Write(buf.Bytes())
		if err != nil || written != buf.Len() {
			log.Errorf("Error writing OP_MSG response: %v, written %d/%d bytes\n", err, written, buf.Len())
		}
	} else {
		log.Errorf("Unsupported requestOpCode: %d", requestOpCode)
	}
}
