package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var checkType2String = map[int]string{
	IS_ID_EXIST:             "数值存在性检查",
	IS_DUP:                  "重复值检查",
	IS_IN_RANGE:             "数值范围检查",
	IS_FILE_EXIST:           "文件存在性检查",
	IS_EMPTY:                "空值检查",
	IS_TIME_CORRECT:         "时间正确性检查",
	IS_IDS_EXIST:            "IDS存在性检查",
	IS_REPEAT_CORRECT:       "Repeat检查",
	IS_SERVER_GROUP_OVERLAP: "服务器分组检查",
	IS_TIME_RANGE_CORRECT:   "时间范围检查",
	LOOT:     "掉落",
	GACHA:    "Gacha",
	TEAMBOSS: "组队BOSS",
}

var errorCode2String = map[int]string{
	NONE: "",
	DATA_UNEXPECTED:                 "数据不符合预期",
	DATA_SOURCE_MISSING:             "来源数据错误",
	DATA_CHECK_TYPE_UNKNOWN:         "未知的检查类型",
	DATA_CONTAINS_INVALID_SEPARATOR: "存在全角分号、逗号或冒号",
	HOT_ACTIVITY_TIME_INVALID:       "活动时间不正确",
	HOT_ACTIVITY_TIME_RANGE_INVALID: "活动时间范围冲突",
	HOT_ACTIVITY_SEVER_ID_INVALID:   "服务器id冲突",
}

type Reporter struct {
	GitStatus        []string
	Unexceptions     []Unexception
	ReportPath       string
	CheckType2String map[int]string
	ErrorCode2String map[int]string
	lastlog          string
}

type Unexception struct {
	Index        int
	ExtraInfo    string
	CheckType    int
	Unexpections []string //内容错误
	ErrorType    int      //错误类型
}

func NewReporter() *Reporter {
	reporter := &Reporter{
		GitStatus:  make([]string, 0, 4),
		ReportPath: filepath.Join(GetVCSRootPath(), "/tools/dataChecker/log"),
	}
	reporter.CheckType2String = checkType2String
	reporter.ErrorCode2String = errorCode2String

	reporter.GetGitStatus()
	return reporter
}

// GetGitStatus汇总并获取git信息
func (r *Reporter) GetGitStatus() {
	gitLog := ".git/FETCH_HEAD"
	if cfg.Dir.RunOnTeamCity {
		gitLog = "gitHEAD.log"
	}

	// tc
	srvGit := filepath.Join(GetVCSRootPath(), gitLog)
	clientGit := filepath.Join(cfg.Dir.ClientProjectDir, gitLog)
	dataGit := filepath.Join(cfg.Dir.DataProjectDir, gitLog)

	// 本地
	if !cfg.Dir.RunOnTeamCity {
		clientGit = filepath.Join(filepath.Dir(cfg.Dir.ClientProjectDir), gitLog)
		dataGit = filepath.Join(filepath.Dir(cfg.Dir.DataProjectDir), gitLog)
	}

	for _, git := range [3]string{srvGit, clientGit, dataGit} {
		r.ReadGitInfo(git)
	}
}

// ReadGitInfo读取单个git的信息
func (r *Reporter) ReadGitInfo(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer file.Close()

	content := make([]string, 4)
	rd := bufio.NewReader(file)
	for {
		str, err := rd.ReadString('\n')
		if str != "" {
			content = append(content, str)
		}
		if err == io.EOF {
			r.GitStatus = append(r.GitStatus, strings.Join(content, ""))
			return
		}
	}
}

// Record 依次记录如下内容：Checklist行号，来源文件名，检查类型，错误类型，错误LOGs
func (r *Reporter) Record(indexOfChecklist int, info string, checkType int, errType int, unexpections []string) {
	ue := Unexception{
		Index:        indexOfChecklist,
		ExtraInfo:    info,
		CheckType:    checkType,
		ErrorType:    errType,
		Unexpections: unexpections,
	}
	r.Unexceptions = append(r.Unexceptions, ue)
}

// Report 生成md格式的报告
func (r *Reporter) Report() {
	if len(r.Unexceptions) == 0 {
		fmt.Println("No error found! Perfect!!!")
		return
	}

	timeSuffix := time.Now().Format("-20060102-150405")
	logFile := filepath.Join(r.ReportPath, "Report"+timeSuffix+".md")
	f, err := os.Create(logFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r.lastlog = logFile

	w := bufio.NewWriter(f)

	w.WriteString("###测试时间：" + time.Now().String() + "\n")

	// 写入Git信息
	w.WriteString("###当前各工程Git Hash：\n")
	for _, gitStat := range r.GitStatus {
		w.WriteString(strings.Join([]string{"* ", gitStat, "\n"}, ""))
	}

	// 写入每条错误信息
	for _, unexpection := range r.Unexceptions {
		r.write(w, unexpection)
	}

	w.Flush()
}

func (r *Reporter) write(w *bufio.Writer, ue Unexception) {
	var idx string

	// 如果没有index则不显示
	if ue.Index != -1 {
		idx = fmt.Sprintf("%d", ue.Index)
	}

	// Title
	line := fmt.Sprintf("### %v %v %v %v：\n",
		r.CheckType2String[ue.CheckType],
		idx,
		ue.ExtraInfo,
		r.ErrorCode2String[ue.ErrorType],
	)
	w.WriteString(line)

	// Content
	if len(ue.Unexpections) > 50 {
		w.WriteString("--== 错误过多，仅显示前50行 ==--\n")
	}

	for i, info := range ue.Unexpections {
		if i == 50 {
			break
		}
		w.WriteString(strings.Join([]string{"* ", info, "\n"}, ""))
	}
}

// DebugRemoveLastlog 调试接口，用于单元测试删除遗留的log文件
func (r *Reporter) DebugRemoveLastlog() {
	if r.lastlog != "" {
		os.Remove(r.lastlog)
	}
}
