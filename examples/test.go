package main

import (
    "github.com/cuixin/log4g"
    "os"
    "time"
)

func main() {
    o, err := os.OpenFile("logging.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
    if err != nil {
        print(err.Error())
        return
    }
    log4g.InitLogger(log4g.LDebug, o)
    defer log4g.Close()
    logBytes := make([]byte, 26)
    for i := 0; i < 26; i++ {
        logBytes[i] = 'a' + byte(i)
    }
    logString := string(logBytes)
    start := time.Now()
    for i := 0; i < 250000; i++ {
        log4g.Debug(logString)
        log4g.Info(logString)
        log4g.Error(logString)
        log4g.Fatal(logString)
    }

    log4g.Info(time.Since(start))
}
