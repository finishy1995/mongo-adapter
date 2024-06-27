package query

import (
	"fmt"
	"github.com/finishy1995/mongo-adapter/library/log"
	"github.com/finishy1995/mongo-adapter/library/tools"
	"github.com/xdg-go/scram"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

const (
	IsMasterCMD  = "ismaster"
	GetNonceCMD  = "getnonce"
	PingCMD      = "ping"
	LogoutCMD    = "logout"
	MechanismCMD = "mechanism"
)

var (
	adminCmdMap = map[string]CmdFunc{}
	server      *scram.Server
)

func init() {
	RegisterAdminCmd(IsMasterCMD, isMaster)
	RegisterAdminCmd(GetNonceCMD, getNonce)
	RegisterAdminCmd(PingCMD, ping)
	RegisterAdminCmd(MechanismCMD, mechanism)
	RegisterAdminCmd(LogoutCMD, logout)
}

func RegisterAdminCmd(cmd string, f CmdFunc) {
	adminCmdMap[cmd] = f
}

func RegisterAuthentication(username string, password string, salt string) error {
	client, err := scram.SHA1.NewClient(username, password, "")
	if err != nil {
		return err
	}
	stored := client.GetStoredCredentials(scram.KeyFactors{
		Salt:  salt,
		Iters: 4096,
	})
	server, err = scram.SHA1.NewServer(func(s string) (scram.StoredCredentials, error) {
		if s == username {
			return stored, nil
		}
		return scram.StoredCredentials{}, fmt.Errorf("user not found")
	})
	if err != nil {
		return err
	}
	return nil
}

func MustRegisterAuthentication(username string, password string, salt string) {
	err := RegisterAuthentication(username, password, salt)
	if err != nil {
		panic(err)
	}
}

func adminFunc(query *OpQuery) (bson.M, error) {
	for key, f := range adminCmdMap {
		if _, ok := query.Query[key]; ok {
			return f(query)
		}
	}

	return nil, fmt.Errorf("unknown admin command: %+v", query.Query)
}

func isMaster(_ *OpQuery) (bson.M, error) {
	return bson.M{
		"ismaster":       true,
		"minWireVersion": 0,
		"maxWireVersion": 8,
		"ok":             1,
		"setName":        "rs0",
	}, nil
}

func getNonce(_ *OpQuery) (bson.M, error) {
	return bson.M{
		"nonce": tools.GetRandomString(16),
		"ok":    1,
	}, nil
}

func ping(_ *OpQuery) (bson.M, error) {
	return bson.M{
		"ok": 1,
	}, nil
}

func logout(_ *OpQuery) (bson.M, error) {
	return bson.M{
		"ok": 1,
	}, nil
}

var (
	resMap = map[string]*scram.ServerConversation{}
)

func mechanism(query *OpQuery) (bson.M, error) {
	if query.Query[MechanismCMD] != "SCRAM-SHA-1" {
		return nil, fmt.Errorf("unsupported mechanism: %s", query.Query[MechanismCMD])
	}

	if server == nil {
		return nil, fmt.Errorf("server is not initialized")
	}

	// 解析 Query，获取 username 和 password
	payload := query.Query["payload"].(primitive.Binary)
	payloadString := string(payload.Data)
	if _, ok := query.Query["saslStart"]; ok {
		conv := server.NewConversation()
		res, err := conv.Step(payloadString)
		if err != nil {
			return nil, err
		}

		conversationId := query.Header.RequestID
		index := strings.Index(res, ",")
		rValue := res[:index]
		resMap[rValue] = conv
		return bson.M{
			"payload": primitive.Binary{
				Subtype: 0,
				Data:    []byte(res),
			},
			"ok":             1,
			"code":           0,
			"done":           false,
			"conversationId": conversationId,
		}, nil
	} else if _, ok := query.Query["saslContinue"]; ok {
		var conv *scram.ServerConversation
		arr := strings.Split(payloadString, ",")
		if len(arr) < 3 {
			return nil, fmt.Errorf("invalid payload: %s", payloadString)
		}
		rValue := arr[1]
		conv = resMap[rValue]
		if conv == nil {
			return nil, fmt.Errorf("conversation not found")
		}

		res, err := conv.Step(payloadString)
		log.Debugf("SCRAM-SHA-1 response: %s", res)
		if err != nil {
			// TODO: fixme, maybe due to mgo too old scram driver
			if res != "e=invalid-proof" {
				return nil, err
			}
		}
		if conv.Done() {
			delete(resMap, rValue)
		}

		return bson.M{
			"done":           true,
			"conversationId": query.Header.RequestID,
			"code":           0,
			"payload": primitive.Binary{
				Subtype: 0,
				Data:    []byte(res),
			},
			"ok": 1,
		}, nil
	} else {
		return nil, fmt.Errorf("unsupported mechanism: %s", query.Query[MechanismCMD])
	}
}
