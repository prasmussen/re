re
==


## Overview
re is a command line utility for pattern matching similar to 'grep', but supports capture groups and multiline matches.
The syntax of the regular expressions accepted is the same general syntax used by Perl, Python, and other languages.
More precisely, it is the syntax accepted by RE2 and described at http://code.google.com/p/re2/wiki/Syntax, except for \C.

## Prerequisites
None, binaries are statically linked.
If you want to compile from source you need the go toolchain: http://golang.org/doc/install

## Installation
- Save the 're' binary to a location in your PATH (i.e. `/usr/local/bin/`)
- Or compile it yourself `go build re.go`


## Usage
    re [options] PATTERN [FILE...]

#### Options
    -d      Delimiter used to seperate capture groups. Default: ", "
    -g      Allow . to match newline (Note: This will read the entire input into memory)
    -i      Ignore case

## Examples
###### "grep mode"
    $ uptime | re "average"
    20:19:29 up 119 days, 23:09,  1 user,  load average: 1.66, 1.56, 1.58

###### Capture group
    $ uptime | re "average: (.+)"
    1.66, 1.56, 1.58

###### Named capture groups
    $ uptime | re "(?P<1min>\d+\.\d+), (?P<5min>\d+\.\d+), (?P<15min>\d+\.\d+)"
    1min=1.66, 5min=1.56, 15min=1.58

###### Named capture groups with custom delimiter
    $ uptime | re -d " -> " "(?P<1min>\d+\.\d+), (?P<5min>\d+\.\d+), (?P<15min>\d+\.\d+)"
    1min=1.66 -> 5min=1.56 -> 15min=1.58

###### Multiline match
    $ ifconfig | re -g "(?P<IF>eth\d+).+?inet addr:(?P<IP>[\d.]+)"
    IF=eth0, IP=10.0.0.100, IF=eth1, IP=10.0.0.101

###### Substitution
    $ uptime | re "s/up/uptime:/"
    20:19:29 uptime: 119 days, 23:09,  1 user,  load average: 1.66, 1.56, 1.58

###### Substitution with capture group
    $ uptime | re "s/(up)/\${1}time:/"
    20:19:29 uptime: 119 days, 23:09,  1 user,  load average: 1.66, 1.56, 1.58