goxxx  [![Build Status](https://travis-ci.org/vaz-ar/goxxx.svg)](https://travis-ci.org/vaz-ar/goxxx)
=====

IRC bot written in Go.

Install
=======

Once you have a working installation of Go, you just need to run:

```
$ go get github.com/vaz-ar/goxxx
```

Build
=======
```
$ make
```

Usage
=====

To get help about program usage, just run:
```
$ goxxx
```

### Configuration file
- By default goxxx will search for a file named `goxxx.ini` in the directory where it is started.
- You can also specify a path for the configuration file via the `-config` flag.

### Log file
- The log file will be created in the directory where goxxx is started, and will be named `goxxx_logs.txt`.

Tests
=====

to run the tests:
```
$ make test
```
It will run the tests for all the packages.


Development / Contributions
=====

Pull requests are welcome.
