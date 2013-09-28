package log4g

import (
    "fmt"
    "io"
    "os"
    "runtime"
    "sync"
    "time"
)

var (
    stdFlags  = lstdFlags | lshortfile
    loggerStd = &logger{out: os.Stderr, flag: stdFlags}
    recvBytes chan []byte
    recvOver  = make(chan bool)
)

var (
    Debugf logfType
    Debug  logType
    Infof  logfType
    Info   logType
    Errorf logfType
    Error  logType
    Fatalf logfType
    Fatal  logType
)

const (
    infoStr  = "[Info ]- "
    debugStr = "[Debug]- "
    errorStr = "[Error]- "
    fatalStr = "[Fatal]- "
)

const (
    LDebug = 1 << iota
    LInfo
    LError
    LFatal
)
const (
    _debug = 1<<(iota+1) - 1
    _info
    _error
    _fatal
)
const (
    ldate = 1 << iota
    ltime
    lmicroseconds
    llongfile
    lshortfile
    lstdFlags = ldate | ltime
)

type logfType func(format string, v ...interface{})
type logType func(v ...interface{})

func Close() {
    close(recvBytes)
    <-recvOver
}

func InitLogger(lvl int, w io.Writer) {
    if w != nil {
        recvBytes = make(chan []byte, 100)
        go func() {
            for v := range recvBytes {
                w.Write(v)
            }
            recvBytes = nil
            recvOver <- true
        }()
    }

    if lvl&_debug != 0 {
        Debugf, Debug = makeLog(debugStr)
    } else {
        Debugf, Debug = emptyLogf, emptyLog
    }
    if lvl&_info != 0 {
        Infof, Info = makeLog(infoStr)
    } else {
        Infof, Info = emptyLogf, emptyLog
    }
    if lvl&_error != 0 {
        Errorf, Error = makeLog(errorStr)
    } else {
        Errorf, Error = emptyLogf, emptyLog
    }
    if lvl&_fatal != 0 {
        Fatalf, Fatal = makeLog(fatalStr)
    } else {
        Fatalf, Fatal = emptyLogf, emptyLog
    }
}

func makeLog(prefix string) (x logfType, y logType) {
    return func(format string, v ...interface{}) {
            loggerStd.output(prefix, 2, fmt.Sprintf(format, v...))
        },
        func(v ...interface{}) {
            loggerStd.output(prefix, 2, fmt.Sprintln(v...))
        }
}

func emptyLogf(format string, v ...interface{}) {}
func emptyLog(v ...interface{})                 {}

type logger struct {
    mu   sync.Mutex
    flag int
    out  io.Writer
    buf  []byte
}

func itoa(buf *[]byte, i int, wid int) {
    var u uint = uint(i)
    if u == 0 && wid <= 1 {
        *buf = append(*buf, '0')
        return
    }

    var b [32]byte
    bp := len(b)
    for ; u > 0 || wid > 0; u /= 10 {
        bp--
        wid--
        b[bp] = byte(u%10) + '0'
    }
    *buf = append(*buf, b[bp:]...)
}

func (l *logger) formatHeader(buf *[]byte, t time.Time, file string, line int) {
    if l.flag&(ldate|ltime|lmicroseconds) != 0 {
        if l.flag&ldate != 0 {
            year, month, day := t.Date()
            itoa(buf, year, 4)
            *buf = append(*buf, '/')
            itoa(buf, int(month), 2)
            *buf = append(*buf, '/')
            itoa(buf, day, 2)
            *buf = append(*buf, ' ')
        }
        if l.flag&(ltime|lmicroseconds) != 0 {
            hour, min, sec := t.Clock()
            itoa(buf, hour, 2)
            *buf = append(*buf, ':')
            itoa(buf, min, 2)
            *buf = append(*buf, ':')
            itoa(buf, sec, 2)
            if l.flag&lmicroseconds != 0 {
                *buf = append(*buf, '.')
                itoa(buf, t.Nanosecond()/1e3, 6)
            }
            *buf = append(*buf, ' ')
        }
    }
    if l.flag&(lshortfile|llongfile) != 0 {
        if l.flag&lshortfile != 0 {
            short := file
            for i := len(file) - 1; i > 0; i-- {
                if file[i] == '/' {
                    short = file[i+1:]
                    break
                }
            }
            file = short
        }
        *buf = append(*buf, file...)
        *buf = append(*buf, ':')
        itoa(buf, line, -1)
        *buf = append(*buf, ": "...)
    }
}

func (l *logger) output(prefix string, calldepth int, s string) error {
    now := time.Now()
    var file string
    var line int
    l.mu.Lock()
    defer l.mu.Unlock()
    if l.flag&(lshortfile|llongfile) != 0 {
        l.mu.Unlock()
        var ok bool
        _, file, line, ok = runtime.Caller(calldepth)
        if !ok {
            file = "???"
            line = 0
        }
        l.mu.Lock()
    }
    l.buf = l.buf[:0]
    l.buf = append(l.buf, prefix...)

    l.formatHeader(&l.buf, now, file, line)
    l.buf = append(l.buf, s...)
    if len(s) > 0 && s[len(s)-1] != '\n' {
        l.buf = append(l.buf, '\n')
    }
    _, err := l.out.Write(l.buf)
    newSlice := make([]byte, len(l.buf))
    if recvBytes != nil {
        copy(newSlice, l.buf)
        recvBytes <- newSlice
    }
    return err
}
