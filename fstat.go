
/*

fstat.go
-John Taylor
Mar 2019

Get info for a list of files across multiple directories

To compile:
go build -ldflags="-s -w" fstat.go

MIT License; Copyright (c) 2018 John Taylor
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

/* TODO

possibly use go channels to improve performance
append optional summary at end in separate table
add option for milliseconds for time stamps

*/


package main

import (
    "bufio"
    "fmt"
    "flag"
    "os"
    "sort"
    "strings"
    "time"

    "github.com/olekukonko/tablewriter"
)

const version = "1.1.0"

type FileStat struct {
    Name string
    FullName string
    Size int64
    ModTime time.Time
    FileType string
}


func sortSize(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if entry[i].Size > entry[j].Size {
            return !ascending
        }
        if entry[i].Size < entry[j].Size {
            return ascending
        }
        // when multiple lines have the same Size, then alphabetize these lines
        return entry[i].Name < entry[j].Name
    })
}

func sortModTime(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if entry[i].ModTime.After(entry[j].ModTime) {
            return !ascending
        }
        if entry[i].ModTime.Before(entry[j].ModTime) {
            return ascending
        }
        // when multiple lines have the same Size, then alphabetize these lines
        return entry[i].Name < entry[j].Name
    })
}

func sortName(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if ascending {
            return entry[i].FullName < entry[j].FullName
        } else {
            return entry[i].FullName > entry[j].FullName
        }
    })
}

func sortNameCaseInsensitive(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if ascending {
            return strings.ToLower(entry[i].FullName) < strings.ToLower(entry[j].FullName)
        } else {
            return strings.ToLower(entry[i].FullName) > strings.ToLower(entry[j].FullName)
        }
    })
}

func GetFileInfo(input *bufio.Scanner, quiet bool) ([]FileStat) {
    var allEntries []FileStat
    fname := ""

    for input.Scan() {
        fname = input.Text()
        f,err := os.Lstat(fname)

        if err != nil {
            if !quiet {
                fmt.Fprintf(os.Stderr, "Error: %s\n", err)
            }
            continue
        }

        var ftype string = "?"
        if f.Mode().IsRegular() {
            ftype = "F"
        } else if f.IsDir() {
            ftype = "D"
        } else if f.Mode() & os.ModeSymlink == os.ModeSymlink {
            ftype = "L"
        }

        entry := FileStat{Name: f.Name(), FullName: fname, Size: f.Size(), ModTime: f.ModTime(), FileType: ftype}
        allEntries = append(allEntries, entry)
    }
    return allEntries
}

func RenderAllEntries(allEntries []FileStat, addCommas bool, convertToMiB bool) {

    var allRows [][]string
    var e FileStat
    var fsize string

    for _,e = range allEntries {
        if convertToMiB {
            e.Size /= 1048576
        }
        if addCommas {
            fsize = RenderInteger("#,###.",e.Size)
        } else {
            fsize = fmt.Sprintf("%d",e.Size)
        }
        //row := []string{fmt.Sprintf("%s",e.ModTime)[:19], fmt.Sprintf("%d",e.Size), fmt.Sprintf("%s",e.FileType), e.FullName}
        row := []string{fmt.Sprintf("%s",e.ModTime)[:19], fsize, fmt.Sprintf("%s",e.FileType), e.FullName}
        allRows = append(allRows, row)
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetAutoWrapText(false)
    table.SetHeader([]string{"Mod Time","Size","Type","Name"})
    table.AppendBulk(allRows)
    if len(allRows) > 0 {
        table.Render()
    }
}

func ValidateArgs(argsSortSize *bool, argsSortSizeDesc *bool, argsSortModTime *bool, argsSortModTimeDesc *bool, argsSortName *bool, argsSortNameDesc *bool, argsSortNameCaseInsen *bool, argsSortNameCaseInsenDesc *bool ) {

    count := 0
    if *argsSortSize { count++ }
    if *argsSortSizeDesc { count++ }
    if *argsSortModTime { count++ }
    if *argsSortModTimeDesc { count++ }
    if *argsSortName { count++ }
    if *argsSortNameDesc { count++ }
    if *argsSortNameCaseInsen { count++ }
    if *argsSortNameCaseInsenDesc { count++ }

    if count > 1 {
        fmt.Fprintf(os.Stderr,"Error: only one sorting argument can be given.\n\n")
        os.Exit(2)
    }
}

func SortAllEntries(allEntries []FileStat, argsSortSize *bool, argsSortSizeDesc *bool, argsSortModTime *bool, argsSortModTimeDesc *bool, argsSortName *bool, argsSortNameDesc *bool, argsSortNameCaseInsen *bool, argsSortNameCaseInsenDesc *bool ) {

    if *argsSortSize { sortSize(allEntries,true) }
    if *argsSortSizeDesc { sortSize(allEntries,false) }
    if *argsSortModTime { sortModTime(allEntries,true) }
    if *argsSortModTimeDesc { sortModTime(allEntries,false) }
    if *argsSortName { sortName(allEntries,true) }
    if *argsSortNameDesc { sortName(allEntries,false) }
    if *argsSortNameCaseInsen { sortNameCaseInsensitive(allEntries,true) }
    if *argsSortNameCaseInsenDesc { sortNameCaseInsensitive(allEntries,false) }
}

func main() {
    argsSortSize := flag.Bool("s", false, "sort by file size")
    argsSortSizeDesc := flag.Bool("S", false, "sort by file size, descending")

    argsSortModTime := flag.Bool("d", false, "sort by file modified date")
    argsSortModTimeDesc := flag.Bool("D", false, "sort by file modified date, newest first")

    argsSortName := flag.Bool("n", false, "sort by file name")
    argsSortNameDesc := flag.Bool("N", false, "sort by file name, reverse alphabetical order")

    argsSortNameCaseInsen := flag.Bool("i", false, "case-insensitive sort by file name")
    argsSortNameCaseInsenDesc := flag.Bool("I", false, "case-insensitive sort by file name, reverse alphabetical order")

    argsVersion := flag.Bool("v", false, "show program version and then exit")
    argsQuiet := flag.Bool("q", false, "do not display file errors")
    argsCommas := flag.Bool("c", false, "add comma thousands separator to file sizes")
    argsMebibytes := flag.Bool("m", false, "convert file sizes to mebibytes")

    flag.Parse()
    if *argsVersion {
        fmt.Fprintf(os.Stderr,"version %s\n", version)
        os.Exit(1)
    }

    ValidateArgs(argsSortSize, argsSortSizeDesc, argsSortModTime, argsSortModTimeDesc, argsSortName, argsSortNameDesc, argsSortNameCaseInsen, argsSortNameCaseInsenDesc)
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

    allEntries := GetFileInfo(input, *argsQuiet)
    SortAllEntries(allEntries,argsSortSize, argsSortSizeDesc, argsSortModTime, argsSortModTimeDesc, argsSortName, argsSortNameDesc, argsSortNameCaseInsen, argsSortNameCaseInsenDesc)
    RenderAllEntries(allEntries, *argsCommas, *argsMebibytes)
}

