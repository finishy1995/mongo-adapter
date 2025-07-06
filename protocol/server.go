package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"finishy1995/mongo-adapter/library/log"

	"go.mongodb.org/mongo-driver/bson"
)

type Conn interface {
	io.Reader
	io.Writer
	RemoteAddr() net.Addr
}

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) OnMessage(conn Conn, buf []byte) error {
	var returnErr error = nil
	hookContext := &HookContext{
		Type: HookStart,
	}
	defer func() {
		if hookContext.ID == "" {
			// 包头无法正确加载
			return
		}
		if hookContext.Type == HookStart {
			hookContext.Type = HookReqHandleBefore
			hookContext.ErrMsg = returnErr.Error()
			fireHook(hookContext)
		}
		hookContext.Type = HookEnd
		hookContext.ErrMsg = ""
		fireHook(hookContext)
	}()

	buffer := bytes.NewBuffer(buf)
	if len(buf) < 16 {
		returnErr = fmt.Errorf("Client sent short packet, len: %d", len(buf))
		return returnErr
	}
	header := MsgHeader{}
	if err := binary.Read(buffer, binary.LittleEndian, &header); err != nil {
		returnErr = fmt.Errorf("Error reading header: %v", err)
		return returnErr
	}
	if len(buf) < int(header.MessageLength) {
		returnErr = fmt.Errorf("Client sent less than header.MessageLength: %d", header.MessageLength)
		return returnErr
	}
	hookContext.ID = getHookID(conn, header.RequestID)

	switch header.OpCode {
	case OP_QUERY: // OP_QUERY
		query := OpQuery{Header: header}
		if err := binary.Read(buffer, binary.LittleEndian, &query.Flags); err != nil {
			returnErr = fmt.Errorf("Error reading flags:", err)
			return returnErr
		}

		collectionName, err := readCString(buffer)
		if err != nil {
			returnErr = fmt.Errorf("Error reading collection name:", err)
			return returnErr
		}
		query.FullCollectionName = collectionName
		if err := binary.Read(buffer, binary.LittleEndian, &query.NumberToSkip); err != nil {
			returnErr = fmt.Errorf("Error reading numberToSkip:", err)
			return returnErr
		}
		if err := binary.Read(buffer, binary.LittleEndian, &query.NumberToReturn); err != nil {
			returnErr = fmt.Errorf("Error reading numberToReturn:", err)
			return returnErr
		}
		// ---- 修正：只取第一个 BSON 文档 ----
		raw := buffer.Bytes()
		if len(raw) < 4 {
			returnErr = fmt.Errorf("Query too short, no BSON length")
			return returnErr
		}
		bsonLen := int(binary.LittleEndian.Uint32(raw[:4]))
		if len(raw) < bsonLen {
			returnErr = fmt.Errorf("Query BSON length out of bound, %d < %d", len(raw), bsonLen)
			return returnErr
		}
		query.Query = raw[:bsonLen]
		query.ReturnFieldsSelector = raw[bsonLen:]

		var cmd bson.D
		if err := bson.Unmarshal(query.Query, &cmd); err != nil {
			returnErr = fmt.Errorf("Error unmarshalling query:", err)
			return returnErr
		}
		log.Debugf("Received OP_QUERY requestID: %d, Collection: %s, Message: %+v", header.RequestID, query.FullCollectionName, cmd)
		hookContext.Request = cmd
		hookContext.Type = HookReqHandleBefore
		fireHook(hookContext)
		s.sendResponse(conn, header.RequestID, header.OpCode, messageHandle(cmd, hookContext))
		break
	case OP_MSG:
		msg := OpMsg{Header: header}
		msgLength := int(header.MessageLength)
		if len(buf) < msgLength {
			returnErr = fmt.Errorf("buf too short for OP_MSG")
			return returnErr
		}
		body := buf[16:msgLength] // 只处理当前消息体
		offset := 0

		if len(body) < 4 {
			returnErr = fmt.Errorf("body too short for flags")
			return returnErr
		}
		msg.Flags = binary.LittleEndian.Uint32(body[offset : offset+4])
		offset += 4

		msg.Sections = make([]Section, 0)
		for offset < len(body) {
			kind := body[offset]
			offset++
			if kind == 0 {
				if offset+4 > len(body) {
					returnErr = fmt.Errorf("Not enough bytes for BSON length")
					return returnErr
				}
				docLen := int(binary.LittleEndian.Uint32(body[offset : offset+4]))
				if offset+docLen > len(body) {
					returnErr = fmt.Errorf("BSON out of bounds")
					return returnErr
				}
				docBytes := body[offset : offset+docLen]
				var doc bson.D
				if err := bson.Unmarshal(docBytes, &doc); err != nil {
					returnErr = fmt.Errorf("Error decoding BSON: %v", err)
					return returnErr
				}
				msg.Sections = append(msg.Sections, Section{Kind: 0, Body: doc})
				offset += docLen
			} else if kind == 1 {
				returnErr = fmt.Errorf("Unsupported Kind == 1")
				return returnErr
			} else {
				returnErr = fmt.Errorf("Unsupported Kind == %d", kind)
				return returnErr
			}
		}
		log.Debugf("Received OP_MSG requestID: %d, Message: %+v", header.RequestID, msg)
		hookContext.Request = msg.Sections[0].Body
		hookContext.Type = HookReqHandleBefore
		fireHook(hookContext)
		s.sendResponse(conn, header.RequestID, header.OpCode, messageHandle(msg.Sections[0].Body, hookContext))
		break
	default:
		returnErr = fmt.Errorf("Received unsupported OpCode: %d\n", header.OpCode)
		return returnErr
	}

	return nil
}

func (s *Server) sendResponse(conn Conn, requestID int32, requestOpCode int32, responseDoc bson.M) {
	log.Debugf("sendResponse. requestID: %d, responseDoc: %+v", requestID, responseDoc)
	hookContext := &HookContext{
		ID:       getHookID(conn, requestID),
		Type:     HookRespHandleAfter,
		Response: responseDoc,
	}
	var err error
	defer func() {
		hookContext.ErrMsg = err.Error()
		fireHook(hookContext)
	}()

	responseBytes, err := bson.Marshal(responseDoc)
	if err != nil {
		log.Errorf("Error marshaling response: %v", err)
		return
	}

	buf := new(bytes.Buffer)
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

		binary.Write(buf, binary.LittleEndian, &header)
		binary.Write(buf, binary.LittleEndian, &reply.ResponseFlags)
		binary.Write(buf, binary.LittleEndian, &reply.CursorID)
		binary.Write(buf, binary.LittleEndian, &reply.StartingFrom)
		binary.Write(buf, binary.LittleEndian, &reply.NumberReturned)
		buf.Write(responseBytes)
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

		binary.Write(buf, binary.LittleEndian, &header)
		binary.Write(buf, binary.LittleEndian, flags) // Flags (4 bytes)
		buf.WriteByte(section0Kind)                   // Section 0 Kind (1 byte)
		buf.Write(responseBytes)                      // Section 0 Body (bson)
	} else {
		log.Errorf("Unsupported requestOpCode: %d", requestOpCode)
		err = fmt.Errorf("Unsupported requestOpCode: %d", requestOpCode)
		return
	}

	written, err := conn.Write(buf.Bytes())
	if err != nil || written != buf.Len() {
		log.Errorf("Error writing response: %v, written %d/%d bytes\n", err, written, buf.Len())
	}
	return
}

func getHookID(conn Conn, requestID int32) string {
	return fmt.Sprintf("%s--%d", conn.RemoteAddr().String(), requestID)
}
