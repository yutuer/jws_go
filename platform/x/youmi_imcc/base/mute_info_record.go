package base

import (
	"fmt"
	"os"

	"time"

	"strconv"

	"github.com/astaxie/beego/utils"
	"github.com/tealeg/xlsx"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

type MuteInfo struct {
	AcID       string
	Reason     string
	IsMute     bool
	MuteTime   int64
	MsgContent string
}

func GenMuteInfo(acID string, reason string, muteTime int64, msgContent string, isMute bool) *MuteInfo {
	return &MuteInfo{
		AcID:       acID,
		Reason:     reason,
		MuteTime:   muteTime,
		MsgContent: msgContent,
		IsMute:     isMute,
	}
}

// simple implementation
type MuteInfoRecorder struct {
	f              *os.File
	lastRecordTime time.Time
	file           *xlsx.File
	sheet          *xlsx.Sheet
}

func (mir *MuteInfoRecorder) Start() {
	mir.openF(time.Now())
	mir.lastRecordTime = time.Now()
}

func (mir *MuteInfoRecorder) CheckSameDay(nowT time.Time) {
	if !mir.isSameDay(mir.lastRecordTime, nowT) {
		mir.closeF()
		mir.SendMail()
		mir.openF(nowT)
	}
}

func (mir *MuteInfoRecorder) Record(muteInfo *MuteInfo, nowT time.Time) {
	mir.CheckSameDay(nowT)
	row := mir.sheet.AddRow()
	values := []string{
		nowT.Format("2006-01-02 15:04:05"),
		muteInfo.AcID,
		muteInfo.Reason,
		strconv.FormatInt(muteInfo.MuteTime, 10),
		muteInfo.MsgContent,
		fmt.Sprintf("%v", muteInfo.IsMute)}
	for _, item := range values {
		cell := row.AddCell()
		cell.SetValue(item)

	}
	mir.lastRecordTime = nowT
}

func (mir *MuteInfoRecorder) SendMail() {
	mail := utils.NewEMail(Cfg.SmtpInfo)
	mail.To = Cfg.SendTo
	mail.From = Cfg.SendFrom
	mail.Subject = fmt.Sprintf("hmt gag list %d-%d-%d",
		mir.lastRecordTime.Year(), mir.lastRecordTime.Month(), mir.lastRecordTime.Day())
	mail.HTML = fmt.Sprintf("<label>Hi，附件中为%d.%d.%d港澳台自动禁言名单, 请核实信息进行封禁和解封，谢谢！</label>",
		mir.lastRecordTime.Year(), mir.lastRecordTime.Month(), mir.lastRecordTime.Day())
	_, err := mail.AttachFile(mir.genRecordFileName(mir.lastRecordTime))
	if err != nil {
		logs.Error("attach file err by %v", err)
	}
	err = mail.Send()
	if err != nil {
		logs.Error("send record mail faild by %v", err)
	}
	logs.Debug("send mail success")
}

func (mir *MuteInfoRecorder) Stop() {
	mir.closeF()
}

func (mir *MuteInfoRecorder) closeF() {
	err := mir.file.Save(mir.genRecordFileName(mir.lastRecordTime))
	if err != nil {
		logs.Error("close file err by %v", err)
	}
}

func (mir *MuteInfoRecorder) openF(now time.Time) {
	var file *xlsx.File
	fileName := mir.genRecordFileName(now)
	file, err := xlsx.OpenFile(fileName)
	if err != nil {
		file = xlsx.NewFile()
	}
	sheet, ok := file.Sheet["MuteInfo"]
	if !ok {
		sheet, err = file.AddSheet("MuteInfo")
		if err != nil {
			logs.Error("add sheet err by %v", err)
		}
		row := sheet.AddRow()
		row.AddCell().SetValue("Time")
		row.AddCell().SetValue("AcID")
		row.AddCell().SetValue("Reason")
		row.AddCell().SetValue("MuteTime")
		row.AddCell().SetValue("Content")
		row.AddCell().SetValue("IsMute")
	}
	mir.file = file
	mir.sheet = sheet
}

func (mir *MuteInfoRecorder) isSameDay(t1, t2 time.Time) bool {
	return t1.Year() == t2.Year() && t1.Month() == t2.Month() && t1.Day() == t2.Day()
}

func (mir *MuteInfoRecorder) genRecordFileName(time time.Time) string {
	return fmt.Sprintf("record/%s_%d_%d_%d.xlsx", recordFilePath, time.Year(),
		time.Month(), time.Day())
}
