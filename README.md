# extract-hamster

extract hamster is a tool to easily extract daily status in gnome hamster to a csv format so it's easier to use this information into a sheet.


## Run

The following shows installing, executing and copying the output to clipboard (xclip)

``` console
$ go install github.com/BrunoTeixeira1996/extract-hamster@latest
$ extract-hamster -range <date range> -out | xclip -selection clipboard
```

The following calcs the output in minutes

``` console
$ extract-hamster -range <date range> -calc-minutes
```

The following outputs only the filtered project (activity in hamster)

``` console
$ extract-hamster -range <date range> -project "PROJECT_HERE"
```
