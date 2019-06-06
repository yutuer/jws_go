package client

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"io/ioutil"

	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"compress/gzip"

	"github.com/timesking/seelog"
	"github.com/ugorji/go/codec"
	"vcs.taiyouxi.net/platform/planx/util/config"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

var mhr codec.MsgpackHandle
var mhw codec.MsgpackHandle

var pLock sync.Mutex
var pLogger seelog.LoggerInterface

type LogPacket struct {
	Timestamp int64
	AccountID string
	RawPacket *Packet
}

func init() {
	mhr.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mhr.RawToString = true
	mhr.WriteExt = true

	mhw.MapType = reflect.TypeOf(map[string]interface{}(nil))
	mhw.RawToString = false
	mhw.WriteExt = true
}

func LoadPacketLogger(lcfgname string, cmd config.LoadCmd) {
	pLock.Lock()
	defer pLock.Unlock()

	switch cmd {
	case config.Load, config.Reload:
		logger, err := seelog.LoggerFromParamConfigAsFile(lcfgname, nil)
		if err != nil {
			logs.Warn("client.LoadPacketLogger load error, %s", err.Error())
			return
		}
		if pLogger == nil {
			pLogger = logger
		} else {
			pLogger.Flush()
			pLogger.Close()
			pLogger = logger
		}
		//fo = NewFileOutput("pakcets.bin")

	case config.Unload:
		if pLogger != nil {
			pLogger.Flush()
			pLogger.Close()
		}
		pLogger = nil

	}
}

func StopPacketLogger() {
	pLock.Lock()
	defer pLock.Unlock()
	if pLogger != nil {
		pLogger.Flush()
		pLogger.Close()
	}
}

func pktToGobBase64(pkt *Packet) string {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	if err := enc.Encode(pkt); err != nil {
		logs.Warn("pktToGobBase64 err: %s", err.Error())
		return ""
	}
	return base64.StdEncoding.EncodeToString(data.Bytes())
}

func PktFromGobBase64(gobBase64 string) *Packet {
	bGob, err := base64.StdEncoding.DecodeString(gobBase64)
	if err != nil {
		logs.Warn("pktFromGobBase64 err: %s", err.Error())
		return nil
	}
	rdata := bytes.NewReader(bGob)
	dec := gob.NewDecoder(rdata)
	var pkt Packet
	if err := dec.Decode(&pkt); err != nil {
		logs.Warn("pktFromGobBase64 err2: %s", err.Error())
		return nil
	}
	return &pkt
}

func pktToJson(pkt *Packet) (action, value string, mvalue map[string]interface{}) {
	var rawBytes []byte

	dec := codec.NewDecoderBytes(pkt.GetBytes(), &mhr)
	dec.Decode(&action)
	dec.Decode(&rawBytes)
	if strings.HasPrefix(action, "/gzip/") {
		action = strings.Replace(action, "/gzip/", "", -1)
		var b bytes.Buffer
		b.Write(rawBytes)
		r, err := gzip.NewReader(&b)
		if err != nil {
			value = fmt.Sprintf("%v", err)
		}
		defer r.Close()
		s, err := ioutil.ReadAll(r)
		if err == nil {
			rawBytes = s
		} else {
			value = fmt.Sprintf("%v", err)
		}

	}
	mapdec := codec.NewDecoderBytes(rawBytes, &mhr)
	mapdec.Decode(&mvalue)
	data, err := json.Marshal(mvalue)

	if err != nil {
		value = fmt.Sprintf("%v", mvalue)
	} else {
		value = string(data)
	}

	return
}

func PktToMap(pkt *Packet) (action string, mvalue map[string]interface{}) {
	var rawBytes []byte

	dec := codec.NewDecoderBytes(pkt.GetBytes(), &mhr)
	dec.Decode(&action)
	dec.Decode(&rawBytes)
	mapdec := codec.NewDecoderBytes(rawBytes, &mhr)
	mapdec.Decode(&mvalue)

	return
}

func jsonToPkt(pktID PacketID, action, jsonv string) *Packet {
	var (
		evalue map[string]interface{}
	)
	err := json.Unmarshal([]byte(jsonv), &evalue)
	if err == nil {
		var out, pktBytes []byte
		enc := codec.NewEncoderBytes(&out, &mhw)
		enc.Encode(evalue)

		encPkt := codec.NewEncoderBytes(&pktBytes, &mhw)
		encPkt.Encode(action)
		encPkt.Encode(out)
		newPkt := NewPacket(pktBytes, pktID)
		return newPkt
	}
	logs.Error("JsonToPkt err pkdID %v action %s json %s", pktID, action, jsonv)
	return nil
}

func MapToPkt(pktID PacketID, action string, mvalue map[string]interface{}) *Packet {
	var out, pktBytes []byte
	enc := codec.NewEncoderBytes(&out, &mhw)
	enc.Encode(mvalue)

	encPkt := codec.NewEncoderBytes(&pktBytes, &mhw)
	encPkt.Encode(action)
	encPkt.Encode(out)
	newPkt := NewPacket(pktBytes, pktID)
	return newPkt
}

//RRTime Request or Response Start Time
const lTypeTime = "time"
const lTypeRecord = "record"
const lTypeInit = "init"
const lTypeSession = "session"

type LogHeader struct {
	TimeOut int64  `json:"logtime"`
	TimeKey string `json:"time"` //time_key in td-agent must be string
	LogType string `json:"type"` //time, record, init, session
}

func newLogHeader(nano int64, Type string) LogHeader {
	return LogHeader{
		nano,
		fmt.Sprintf("%d", nano),
		strings.ToLower(Type), //make sure, it is lower case, so
	}
}

type LogTime struct {
	LogHeader
	ResponseTime int64  `json:"dur"`
	RequestURL   string `json:"req"`
	ResponseURL  string `json:"resp"`
}

type LogRecord struct {
	LogHeader
	RType       int    `json:"rtype"`  // Request&Response or Content
	RRPrefix    string `json:"prefix"` //RES or RESP or CONTENT
	Account     string `json:"accountid"`
	Passthrough string `json:"passthrough"`
	Action      string `json:"uri"`
	RawJson     string `json:"pkt"`
	BinaryLen   int    `json:"len"`
	BinaryB64   string `json:"bin"`
}

type LogInitRecord struct {
	LogHeader
	Account  string `json:"accountid"`
	RandSeed int64  `json:"randseed"`
}

type LogSessionRecord struct {
	LogHeader
	Account string `json:"accountid"`
	Session string `json:"session"`
}

type RRTime int64

func LogRequestStartTime() RRTime {
	if pLogger == nil {
		return 0
	}
	return RRTime(time.Now().UnixNano())
}

func LogResponseTime(request, response string, start RRTime) {
	if pLogger == nil {
		return
	}

	lt := LogTime{
		LogHeader: newLogHeader(int64(start), lTypeTime),
	}
	now := time.Now().UnixNano()
	lt.ResponseTime = now - int64(start)
	lt.RequestURL = request
	lt.ResponseURL = response
	out, _ := json.Marshal(&lt)

	pLock.Lock()
	defer pLock.Unlock()
	pLogger.Infof("%s", out)
}

func Log(prefix, accountId string, pkt *Packet) {
	if pLogger == nil {
		return
	}

	lr := LogRecord{
		LogHeader: newLogHeader(time.Now().UnixNano(), lTypeRecord),
	}
	lr.RType = int(pkt.GetContentType())
	lr.Account = accountId
	lr.RRPrefix = strings.ToLower(prefix)

	// by zhangzhen 为了查协议大小问题，精简了packetlog内容，pkt和bin的内容没有记，pingpong也没记
	switch pkt.GetContentType() {
	case PacketIDContent:
		action, _, _ := pktToJson(pkt)
		//action, value, _ := pktToJson(pkt)
		//gobBase64 := pktToGobBase64(pkt)

		lr.Action = action
		//lr.RawJson = value
		lr.BinaryLen = int(pkt.len)
		//lr.BinaryB64 = gobBase64

	case PacketIDReqResp:
		action, _, mvalue := pktToJson(pkt)
		//gobBase64 := pktToGobBase64(pkt)

		pt, ok := mvalue["passthrough"]
		if ok && pt != nil {
			lr.Passthrough = pt.(string)
		} else {
			lr.Passthrough = "WRONGPacket"
			//logs.Error("Packet Logger catch a problem, no passthrough! %s %s %s %s %v %v",
			//	action, value, prefix, accountId, mvalue, pkt.GetBytes())
		}
		lr.Action = action
		//lr.RawJson = value
		lr.BinaryLen = int(pkt.len)
		//lr.BinaryB64 = gobBase64
	}

	out, _ := json.Marshal(&lr)

	pLock.Lock()
	defer pLock.Unlock()

	if lr.RType != int(PacketIDPingPong) {
		pLogger.Infof("%s", out)
	}
}

func LogInitData(accountid string, seed int64) {
	if pLogger == nil {
		return
	}
	//accountid
	lir := LogInitRecord{
		LogHeader: newLogHeader(time.Now().UnixNano(), lTypeInit),
		Account:   accountid,
		RandSeed:  seed,
	}

	out, _ := json.Marshal(&lir)

	pLock.Lock()
	defer pLock.Unlock()
	pLogger.Infof("%s", out)
}

func LogSession(accountid string, session_action string) {
	if pLogger == nil {
		return
	}
	//accountid
	loolr := LogSessionRecord{
		LogHeader: newLogHeader(time.Now().UnixNano(), lTypeSession),
		Account:   accountid,
		Session:   session_action,
	}

	out, _ := json.Marshal(&loolr)

	pLock.Lock()
	defer pLock.Unlock()
	pLogger.Infof("%s", out)

}
