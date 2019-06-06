// Package lumberjack provides a rolling logger.
//
// Note that this is v2.0 of lumberjack, and should be imported using gopkg.in
// thusly:
//
//   import "gopkg.in/natefinch/lumberjack.v2"
//
// The package name remains simply lumberjack, and the code resides at
// https://github.com/natefinch/lumberjack under the v2.0 branch.
//
// Lumberjack is intended to be one part of a logging infrastructure.
// It is not an all-in-one solution, but instead is a pluggable
// component at the bottom of the logging stack that simply controls the files
// to which logs are written.
//
// Lumberjack plays well with any logging package that can write to an
// io.Writer, including the standard library's log package.
//
// Lumberjack assumes that only one process is writing to the output files.
// Using the same lumberjack configuration from multiple processes on the same
// machine will result in improper behavior.
package lumberjack

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//	backupTimeFormat = "2006-01-02T15-04-05.000"
	backupTimeFormat = "2006-01-02"
	defaultMaxSize   = 100
	fileNumSpliter   = "_"
)

// ensure we always implement io.WriteCloser
var _ io.WriteCloser = (*Logger)(nil)

// Logger is an io.WriteCloser that writes to the specified filename.
//
// Logger opens or creates the logfile on first Write.  If the file exists and
// is less than MaxSize megabytes, lumberjack will open and append to that file.
// If the file exists and its size is >= MaxSize megabytes, the file is renamed
// by putting the current time in a timestamp in the name immediately before the
// file's extension (or the end of the filename if there's no extension). A new
// log file is then created using original filename.
//
// Whenever a write would cause the current log file exceed MaxSize megabytes,
// the current file is closed, renamed, and a new log file created with the
// original name. Thus, the filename you give Logger is always the "current" log
// file.
//
// Cleaning Up Old Log Files
//
// Whenever a new logfile gets created, old log files may be deleted.  The most
// recent files according to the encoded timestamp will be retained, up to a
// number equal to MaxBackups (or all of them if MaxBackups is 0).  Any files
// with an encoded timestamp older than MaxAge days are deleted, regardless of
// MaxBackups.  Note that the time encoded in the timestamp is the rotation
// time, which may differ from the last time that file was written to.
//
// If MaxBackups and MaxAge are both 0, no old log files will be deleted.
type Logger struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-lumberjack.log in
	// os.TempDir() if empty.
	FileTempletName string `json:"filetempletname" yaml:"filetempletname"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize"`

	// BufSize is buf writer's size, if 0, then will be default
	BufSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	//	LocalTime bool `json:"localtime" yaml:"localtime"`

	// TimeLocal determines if the time used for formatting the timestamps in
	// log files is TimeLocal time.  The default is to use UTC time.
	TimeLocal string `json:"timelocal" yaml:"timelocal"`

	GetUTCSec GetUTCSec `json:"getutcsec" yaml:"getutcsec"`

	Header string // 文件头，默认""无效

	size int64
	file *os.File

	bufWriter *bufio.Writer

	fileTimeStamp string
	mu            sync.Mutex
	timeLocal     *time.Location
}

var (
	// currentTime exists so it can be mocked out by tests.
	currentTime = time.Now

	// os_Stat exists so it can be mocked out by tests.
	os_Stat = os.Stat

	// megabyte is the conversion factor between MaxSize and bytes.  It is a
	// variable so tests can mock it out and not need to write megabytes of data
	// to disk.
	megabyte = 1024 * 1024
)

type GetUTCSec func() int64

// Write implements io.Writer.  If a write would cause the log file to be larger
// than MaxSize, the file is closed, renamed to include a timestamp of the
// current time, and a new log file is created using the original log file name.
// If the length of the write is greater than MaxSize, an error is returned.
func (l *Logger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	writeLen := int64(len(p))
	if writeLen > l.max() {
		return 0, fmt.Errorf(
			"write length %d exceeds maximum file size %d", writeLen, l.max(),
		)
	}

	if l.file == nil {
		if err = l.openExistingOrNew(len(p)); err != nil {
			return 0, err
		}
	}

	if l.size+writeLen > l.max() || l.getFileTimeStamp() != l.fileTimeStamp {
		if err := l.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = l.bufWriter.Write(p)
	l.size += int64(n)

	return n, err
}

// Close implements io.Closer, and closes the current logfile.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.close()
}

// close closes the file if it is open.
func (l *Logger) close() error {
	if l.file == nil {
		return nil
	}
	l.bufWriter.Flush()
	err := l.file.Close()
	l.file = nil
	return err
}

// Rotate causes Logger to close the existing log file and immediately create a
// new one.  This is a helper function for applications that want to initiate
// rotations outside of the normal rotation rules, such as in response to
// SIGHUP.  After rotating, this initiates a cleanup of old log files according
// to the normal rules.
func (l *Logger) Rotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.rotate()
}

// rotate closes the current file, moves it aside with a timestamp in the name,
// (if it exists), opens a new file with the original filename, and then runs
// cleanup.
func (l *Logger) rotate() error {
	if err := l.close(); err != nil {
		return err
	}

	_, nextFileName, err := l.getFileName()
	if err != nil {
		return err
	}
	if err := l.openNew(nextFileName); err != nil {
		return err
	}
	return l.cleanup()
}

// openNew opens a new log file for writing, moving any old log file out of the
// way.  This methods assumes the file has already been closed.
func (l *Logger) openNew(fileName string) error {
	err := os.MkdirAll(l.dir(), 0744)
	if err != nil {
		return fmt.Errorf("can't make directories for new logfile: %s", err)
	}

	mode := os.FileMode(0644)
	// we use truncate here because this should only get called when we've moved
	// the file ourselves. if someone else creates the file in the meantime,
	// just wipe out the contents.
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("can't open new logfile: %s", err)
	}
	if l.Header != "" {
		if _, err := f.Write([]byte(l.Header)); err != nil {
			return err
		}
	}
	pre, suf := l.prefixAndExt()
	timestamp, _ := l.timeFromName(fileName, pre, suf)
	l.file = f
	l.fileTimeStamp = timestamp
	l.size = 0
	l.bufWriter = bufio.NewWriterSize(l.file, l.BufSize)
	return nil
}

// openExistingOrNew opens the logfile if it exists and if the current write
// would not put it over MaxSize.  If there is no such file or the write would
// put it over the MaxSize, a new file is created.
func (l *Logger) openExistingOrNew(writeLen int) error {
	lastFileName, nextFileName, err := l.getFileName()
	if lastFileName == "" {
		return l.openNew(nextFileName)
	}
	if err != nil {
		return fmt.Errorf("error getting log file info: %s", err)
	}

	info, err := os_Stat(lastFileName)
	if info.Size()+int64(writeLen) >= l.max() {
		return l.rotate()
	}

	file, err := os.OpenFile(lastFileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// if we fail to open the old log file for some reason, just ignore
		// it and open a new log file.
		return l.openNew(nextFileName)
	}
	pre, suf := l.prefixAndExt()
	timestamp, _ := l.timeFromName(lastFileName, pre, suf)
	l.file = file
	l.fileTimeStamp = timestamp
	l.size = info.Size()
	l.bufWriter = bufio.NewWriterSize(l.file, l.BufSize)
	return nil
}

// genFilename generates the name of the logfile from the current time.
func (l *Logger) filetempletname() string {
	if l.FileTempletName != "" {
		return l.FileTempletName
	}
	name := filepath.Base(os.Args[0]) + "-lumberjack.log"
	return filepath.Join(os.TempDir(), name)
}

// cleanup deletes old log files, keeping at most l.MaxBackups files, as long as
// none of them are older than MaxAge.
func (l *Logger) cleanup() error {
	if l.MaxBackups == 0 && l.MaxAge == 0 {
		return nil
	}

	files, err := l.oldLogFiles()
	if err != nil {
		return err
	}

	var deletes []logInfo

	if l.MaxBackups > 0 && l.MaxBackups < len(files) {
		deletes = files[l.MaxBackups:]
		files = files[:l.MaxBackups]
	}
	if l.MaxAge > 0 {
		diff := time.Duration(int64(24*time.Hour) * int64(l.MaxAge))

		cutoff := currentTime().Add(-1 * diff)

		for _, f := range files {
			if f.timestamp.Before(cutoff) {
				deletes = append(deletes, f)
			}
		}
	}

	if len(deletes) == 0 {
		return nil
	}

	go deleteAll(l.dir(), deletes)

	return nil
}

func deleteAll(dir string, files []logInfo) {
	// remove files on a separate goroutine
	for _, f := range files {
		// what am I going to do, log this?
		_ = os.Remove(filepath.Join(dir, f.Name()))
	}
}

// oldLogFiles returns the list of backup log files stored in the same
// directory as the current log file, sorted by ModTime
func (l *Logger) oldLogFiles() ([]logInfo, error) {
	files, err := ioutil.ReadDir(l.dir())
	if err != nil {
		return nil, fmt.Errorf("can't read log file directory: %s", err)
	}
	logFiles := []logInfo{}

	prefix, ext := l.prefixAndExt()

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name, _ := l.timeFromName(f.Name(), prefix, ext)
		if name == "" {
			continue
		}
		t, err := time.ParseInLocation(backupTimeFormat, name, l.getTimeLocal())
		if err == nil {
			logFiles = append(logFiles, logInfo{t, f})
		}
		// error parsing means that the suffix at the end was not generated
		// by lumberjack, and therefore it's not a backup file.
	}

	sort.Sort(byFormatTime(logFiles))

	return logFiles, nil
}

// timeFromName extracts the formatted time from the filename by stripping off
// the filename's prefix and extension. This prevents someone's filename from
// confusing time.parse.
func (l *Logger) timeFromName(filename, prefix, ext string) (string, int) {
	filename = filepath.Base(filename)
	if !strings.HasPrefix(filename, prefix) {
		return "", 0
	}
	filename = filename[len(prefix):]

	if !strings.HasSuffix(filename, ext) {
		return "", 0
	}
	filename = filename[:len(filename)-len(ext)]
	i := strings.Index(filename, fileNumSpliter)
	if i > 0 {
		num, err := strconv.Atoi(filename[i+1:])
		if err != nil {
			return "", 0
		}
		return filename[:i], num
	}
	return filename, 0
}

// max returns the maximum size in bytes of log files before rolling.
func (l *Logger) max() int64 {
	if l.MaxSize == 0 {
		return int64(defaultMaxSize * megabyte)
	}
	return int64(l.MaxSize) * int64(megabyte)
}

// dir returns the directory for the current filename.
func (l *Logger) dir() string {
	return filepath.Dir(l.filetempletname())
}

// prefixAndExt returns the filename part and extension part from the Logger's
// filename.
func (l *Logger) prefixAndExt() (prefix, ext string) {
	filename := filepath.Base(l.filetempletname())
	ext = filepath.Ext(filename)
	prefix = filename[:len(filename)-len(ext)] + "_"
	return prefix, ext
}

func (l *Logger) getFileName() (lastFileName, nextFileName string, err error) {
	files, err := ioutil.ReadDir(l.dir())
	if err != nil {
		return "", "", fmt.Errorf("can't read log file directory: %s", err)
	}

	prefix, ext := l.prefixAndExt()
	timestamp := l.getFileTimeStamp()
	numMax := -1
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name, n := l.timeFromName(f.Name(), prefix, ext)
		if name == "" {
			continue
		}
		if name == timestamp {
			if n > numMax {
				numMax = n
				lastFileName = f.Name()
			}
		}
	}
	fileDir := filepath.Dir(l.filetempletname())
	if lastFileName != "" {
		lastFileName = filepath.Join(fileDir, lastFileName)
	}
	if numMax < 0 {
		return lastFileName, filepath.Join(fileDir, prefix+timestamp+ext), nil
	}
	return lastFileName, filepath.Join(fileDir, prefix+timestamp+fileNumSpliter+fmt.Sprint(numMax+1)+ext), nil
}

func (l *Logger) getFileTimeStamp() string {
	if l.GetUTCSec == nil {
		l.GetUTCSec = getUTCSec
	}
	_t := time.Unix(l.GetUTCSec(), 0)
	t := _t.In(l.getTimeLocal())
	return t.Format(backupTimeFormat)
}

func (l *Logger) getTimeLocal() *time.Location {
	if l.timeLocal == nil {
		tl, err := time.LoadLocation(l.TimeLocal)
		if err != nil {
			panic(err)
		}
		l.timeLocal = tl
	}
	return l.timeLocal
}

func getUTCSec() int64 {
	return time.Now().Unix()
}

// logInfo is a convenience struct to return the filename and its embedded
// timestamp.
type logInfo struct {
	timestamp time.Time
	os.FileInfo
}

// byFormatTime sorts by newest time formatted in the name.
type byFormatTime []logInfo

func (b byFormatTime) Less(i, j int) bool {
	return b[i].timestamp.After(b[j].timestamp)
}

func (b byFormatTime) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byFormatTime) Len() int {
	return len(b)
}
