log4g
=======

Simple go logger, such as log4j, too simple to use.

installation
------------

    go get github.com/cuixin/log4g

usage
-----

Add main entry to start log4g, specify the output file or not(No
output set the nil argument, just only print to console)


```
    import "github.com/cuixin/log4g"

    func main() {
        o, err := os.OpenFile("logging.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
        if err != nil {
            print(err.Error())
            return
        }
        log4g.InitLogger(log4g.LDebug, o)
        defer log4g.Close()
        // begin to output
        log4g.Info("hello world")
    }
```

examples&benchmark
-------
    test 1000k lines to print:
    $>go run examples/test.go
    $>tail -f logging.log

    or directly test the speed of writing bytes.
    ```
    $>nohup go run examples/test.go >/dev/null 2>&1
    $>tail -f logging.log
    ```
