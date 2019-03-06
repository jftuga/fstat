
package main

import (
    "bufio"
    "fmt"
    "flag"
    "os"

    "github.com/olekukonko/tablewriter"
)

func GetFileInfo(input *bufio.Scanner) {
    var allRows [][]string
    fname := ""

    for input.Scan() {
        fname = input.Text()
        f,_ := os.Stat(fname)

        fmt.Printf("%s\t%d\t%s\n", f.ModTime(), f.Size(), f.Name())
        row := []string{fmt.Sprintf("%s",f.ModTime())[:19], fmt.Sprintf("%d",f.Size()), fname}
        allRows = append(allRows, row)
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetAutoWrapText(false)
    table.SetHeader([]string{"Mod Time","Size","Name"})
    table.AppendBulk(allRows)
    table.Render()
}

func main() {
    flag.Parse()
    args := flag.Args()

    var input *bufio.Scanner
    if 0 == len(args) { // read from STDIN
        input = bufio.NewScanner(os.Stdin)
    } else { // read from filename
        fname := args[0]
        file, err := os.Open(fname)
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s\n", err)
            os.Exit(1)
        }
        defer file.Close()
        input = bufio.NewScanner(file)
    }

    GetFileInfo(input)
}

