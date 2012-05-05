package main

import (
    "fmt"
    "regexp"
    "regexp/syntax"
    "os"
    "io"
    "bufio"
    "flag"
    "strings"
)

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

func (self *Re) FindMatches(fnames []string) {
    if len(fnames) == 0 {
        reader := bufio.NewReader(os.Stdin)
        self.printMatches(reader, "")
    } else {
        for _, fname := range fnames {
            f, err := os.Open(fname)
            if err != nil {
                fmt.Fprintln(os.Stderr, err)
                continue
            }
            prefix := ""
            if (len(fnames) > 1) {
                prefix = fname
            }
            reader := bufio.NewReader(f)
            self.printMatches(reader, prefix)
            f.Close()
        }
    }
}

func (self *Re) printMatches(reader *bufio.Reader, prefix string) {
    lines := make([]string, 0)
    var lineBuffer string

    for {
        bytes, hasMore, err := reader.ReadLine()
        if err == io.EOF {
            break
        } else if err != nil {
            fmt.Println(err)
        }

        line := lineBuffer + string(bytes)
        if hasMore {
            lineBuffer = line
            continue
        }

        if self.dotAll {
            lines = append(lines, line)
            continue
        }

        if self.groupCount == 0 && self.re.MatchString(line) {
            fmt.Println(prefix + line)
        } else if self.groupCount > 0 {
            self.printCaptureGroups(line, prefix)
        }
        lineBuffer = ""
    }
}

func (self *Re) printCaptureGroups(line, prefix string) {
    matches := self.re.FindAllStringSubmatch(line, -1)
    if matches == nil {
        return
    }
    entries := make([]string, 0)
    for _, m := range matches {
        groups := m[1:]
        for i, group := range groups {
            var entry string
            if name := self.groupNames[i]; name != "" {
                entry = fmt.Sprintf("%s=%s", name, group)
            } else {
                entry = group
            }
            entries = append(entries, entry)
        }
    }
    fmt.Println(prefix + strings.Join(entries, self.delimiter))
}

func main() {
    ignoreCase := flag.Bool("i", false, "Ignore case")
    dotAll := flag.Bool("g", false, "Allow . to match newline")
    delimiter := flag.String("d", ", ", "Delimiter used to seperate capture groups")
    flag.Parse()

    args := flag.Args()
    if (len(args) == 0) {
        fmt.Fprintln(os.Stderr, "Missing pattern")
        os.Exit(1)
    }

    re, err := NewRe(args[0], *delimiter, *ignoreCase, *dotAll)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    re.FindMatches(args[1:])
}
