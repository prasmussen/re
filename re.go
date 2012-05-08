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

type Line struct {
    Prefix string
    Body string
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

func (self *Re) FindMatches(fnames []string, matches, errors chan string) {
    lines := make(chan *Line)
    go self.patternMatcher(lines, matches)

    var reader func(*Re, *bufio.Reader, chan *Line, string, chan string)
    if self.dotAll {
        reader = (*Re).readAllReader
    } else {
        reader = (*Re).lineReader
    }

    if len(fnames) == 0 {
        reader(self, bufio.NewReader(os.Stdin), lines, "", errors)
    } else {
        for _, fname := range fnames {
            f, err := os.Open(fname)
            if err != nil {
                errors<- fmt.Sprintf("%s", err)
                continue
            }
            var prefix string
            if (len(fnames) > 1) {
                prefix = fname + ":"
            }
            reader(self, bufio.NewReader(f), lines, prefix, errors)
            f.Close()
        }
    }
    close(errors)
    close(lines)
}

func (self *Re) readAllReader(reader *bufio.Reader, lines chan *Line, prefix string, errors chan string) {
    bytes, err := ioutil.ReadAll(reader)
    if err != nil {
        errors<- fmt.Sprintf("%s", err)
        return
    }
    lines<- &Line{prefix, string(bytes)}
}

func (self *Re) lineReader(reader *bufio.Reader, lines chan *Line, prefix string, errors chan string) {
    var lineBuffer string

    for {
        bytes, hasMore, err := reader.ReadLine()
        if err != nil {
            if err != io.EOF {
                errors<- fmt.Sprintf("%s", err)
            }
            break
        }

        line := lineBuffer + string(bytes)
        if hasMore {
            lineBuffer = line
            continue
        }
        lines<- &Line{prefix, line}
        lineBuffer = ""
    }
}

func (self *Re) patternMatcher(lines chan *Line, matches chan string) {
    for line := range lines {
        var match string
        if self.groupCount == 0 && self.re.MatchString(line.Body) {
            match = line.Body
        } else if self.groupCount > 0 {
            match = self.getCaptureGroups(line.Body)
        }

        if match != "" {
            matches<- fmt.Sprintf("%s%s", line.Prefix, match)
        }
    }
    close(matches)
}

func (self *Re) getCaptureGroups(line string) string {
    matches := self.re.FindAllStringSubmatch(line, -1)
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

func printOutput(matches, errors chan string, done chan bool) {
    go func() {
        for err := range errors {
            fmt.Fprintln(os.Stderr, err)
        }
    }()

    for match := range matches {
        fmt.Println(match)
    }
    done<- true
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
    flag.Parse()

    args := flag.Args()
    if len(args) == 0 {
        usage()
    }

    re, err := NewRe(args[0], *delimiter, *ignoreCase, *dotAll)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    matches := make(chan string, 10)
    errors := make(chan string, 10)
    done := make(chan bool)
    go printOutput(matches, errors, done)
    re.FindMatches(args[1:], matches, errors)
    <-done
}


