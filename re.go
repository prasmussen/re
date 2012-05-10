package main

import (
    "fmt"
    "regexp"
    "regexp/syntax"
    "os"
    "io"
    "io/ioutil"
    "bufio"
    "flag"
    "strings"
)

type UnitReaderType int

const (
    LineReader UnitReaderType = iota
    AllReader
)

type IOUnit struct {
    Name string
    Data string
}

func NewIOUnit(name, data string) *IOUnit {
    return &IOUnit{name, data}
}

type FileIO struct {
    unitReader func(*FileIO, *bufio.Reader, string, chan error, chan *IOUnit)
}

func NewFileIO(t UnitReaderType) *FileIO {
    fio := &FileIO{}
    switch t {
        case LineReader:
            fio.unitReader = (*FileIO).lineReader
        case AllReader:
            fio.unitReader = (*FileIO).allReader
    }
    return fio
}

func (self *FileIO) ReadFiles(fnames []string) (chan *IOUnit, chan error) {
    errors := make(chan error)
    units := make(chan *IOUnit)
    go self.fileReader(fnames, errors, units)
    return units, errors
}

func (self *FileIO) fileReader(fnames []string, errors chan error, units chan *IOUnit) {
    defer close(units)
    defer close(errors)

    // Use stdin if no files are provided
    if len(fnames) == 0 {
        self.unitReader(self, bufio.NewReader(os.Stdin), "stdin", errors, units)
        return
    }

    for _, fname := range fnames {
        f, err := os.Open(fname)
        if err != nil {
            errors<- err
            continue
        }
        self.unitReader(self, bufio.NewReader(f), fname, errors, units)
        f.Close()
    }
}

func (self *FileIO) lineReader(reader *bufio.Reader, name string, errors chan error, units chan *IOUnit) {
    var lineBuffer string

    for {
        bytes, hasMore, err := reader.ReadLine()
        if err != nil {
            if err != io.EOF {
                errors<- err
            }
            break
        }

        line := lineBuffer + string(bytes)
        if hasMore {
            lineBuffer = line
            continue
        }
        units<- NewIOUnit(name, line)
        lineBuffer = ""
    }
}

func (self *FileIO) allReader(reader *bufio.Reader, name string, errors chan error, units chan *IOUnit) {
    bytes, err := ioutil.ReadAll(reader)
    if err != nil {
        errors<- err
        return
    }
    units<- NewIOUnit(name, string(bytes))
}

type Result struct {
    Data string
    Unit *IOUnit
}

func NewResult(data string, unit *IOUnit) *Result {
    return &Result{data, unit}
}


type Re struct {
    re *regexp.Regexp
    dotAll bool
    delimiter string
    groupCount int
    groupNames []string
}

func NewRe(pattern, delimiter string, ignoreCase, dotAll bool) (*Re, error) {
    flags := syntax.Perl
    if ignoreCase {
        flags |= syntax.FoldCase
    }
    if dotAll {
        flags |= syntax.DotNL
    }

    meta, err := syntax.Parse(pattern, flags)
    if err != nil {
        return nil, err
    }

    return &Re{
        re: regexp.MustCompile(meta.String()),
        dotAll: dotAll,
        delimiter: delimiter,
        groupCount: meta.MaxCap(),
        groupNames: meta.CapNames()[1:],
    }, nil
}

func (self *Re) Replace(fnames []string, matches, errors chan string) { }

func (self *Re) Find(units chan *IOUnit) (chan *Result) {
    results := make(chan *Result)
    go self.patternMatcher(units, results)
    return results
}

func (self *Re) patternMatcher(units chan *IOUnit, results chan *Result) {
    defer close(results)

    for unit := range units {
        var data string
        if self.groupCount == 0 && self.re.MatchString(unit.Data) {
            // There is a match, but the regex has no capture groups so we set 'output data' = 'input data'
            data = unit.Data
        } else if self.groupCount > 0 {
            // The regex has at least one capture group, if the regex does not match; data will be empty
            data = self.getCaptureGroups(unit.Data)
        }

        if data != "" {
            results<- NewResult(data, unit)
        }
    }
}

// Returns a string with each capture group seperated by self.delimiter
// Returns an empty string if the regex does not match the input data
func (self *Re) getCaptureGroups(data string) string {
    matches := self.re.FindAllStringSubmatch(data, -1)
    if matches == nil {
        return ""
    }
    entries := make([]string, 0)
    for _, m := range matches {
        groups := m[1:]
        for i, group := range groups {
            entries = append(entries, self.prependGroupName(group, i))
        }
    }
    return strings.Join(entries, self.delimiter)
}

func (self *Re) prependGroupName(group string, index int) string {
    if name := self.groupNames[index]; name != "" {
        return fmt.Sprintf("%s=%s", name, group)
    }
    return group
}

func printErrors(errors chan error) {
    for err := range errors {
        fmt.Fprintln(os.Stderr, err)
    }
}

func printResults(results chan *Result, fnamePrefix bool) {
    for result := range results {
        var output string
        if fnamePrefix {
            output = fmt.Sprintf("%s:%s", result.Unit.Name, result.Data)
        } else {
            output = result.Data
        }
        fmt.Println(output)
    }
}

func usage() {
    fmt.Fprintln(os.Stderr, "Usage: re [options] PATTERN [FILE...]")
    flag.PrintDefaults()
    os.Exit(1)
}

func main() {
    ignoreCase := flag.Bool("i", false, "Ignore case")
    dotAll := flag.Bool("g", false, "Allow . to match newline")
    delimiter := flag.String("d", ", ", "Delimiter used to seperate capture groups")
    //replaceMode := flag.Bool("r", false, "Replace mode")
    flag.Parse()

    args := flag.Args()
    if len(args) == 0 {
        usage()
    }
    pattern, fnames := args[0], args[1:]
    prefixFnames := len(fnames) > 1

    re, err := NewRe(pattern, *delimiter, *ignoreCase, *dotAll)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    var readerType UnitReaderType
    if *dotAll {
        readerType = AllReader
    } else {
        readerType = LineReader
    }
    fio := NewFileIO(readerType)

    units, errors := fio.ReadFiles(fnames)
    results := re.Find(units)

    go printErrors(errors)
    printResults(results, prefixFnames)
}


