re -- grep with capture groups!
==============================


## Usage
    re [options] PATTERN [FILE...]

#### Options
    -d      Delimiter used to seperate capture groups. Default: ", "
    -g      Allow . to match newline (Note: This will read the entire file into memory)
    -i      Ignore case

#### Examples
    # "grep mode"
    $ uptime | re "average"
    20:19:29 up 119 days, 23:09,  1 user,  load average: 1.66, 1.56, 1.58

    # Capture group
    $ uptime | re "average: (.+)"
    1.66, 1.56, 1.58

    # Named capture groups
    $ uptime | re "(?P<1min>\d+\.\d+), (?P<5min>\d+\.\d+), (?P<15min>\d+\.\d+)"
    1min=1.66, 5min=1.56, 15min=1.58

    # Named capture groups with custom delimiter
    $ uptime | re -d " -> " "(?P<1min>\d+\.\d+), (?P<5min>\d+\.\d+), (?P<15min>\d+\.\d+)"
    1min=1.66 -> 5min=1.56 -> 15min=1.58

    # Multiline match
    $ ifconfig | re -g "(?P<IF>eth\d+).+?inet addr:(?P<IP>[\d.]+)"
    IF=eth0, IP=10.0.0.100, IF=eth1, IP=10.0.0.101

