# extract-hamster

extract hamster is a tool to easily extract daily status in gnome hamster to a csv format so it's easier to use this information into a sheet.


## Run

The following shows installing, executing and copying the output to clipboard (xclip)

``` console
$ go install github.com/BrunoTeixeira1996/extract-hamster@latest
$ extract-hamster -range <date range> | xclip -selection clipboard
```

