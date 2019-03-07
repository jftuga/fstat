
/*

fstat.go
-John Taylor
Mar 2019

Get info for a list of files across multiple directories

To compile:
go build -ldflags="-s -w" fstat.go render_number.go

MIT License; Copyright (c) 2019 John Taylor
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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

const version = "1.2.1"

type FileStat struct {
    Name string
    FullName string
    Size int64
    ModTime time.Time
    FileType string
}

/*
sortSize sorts the FileStat slice by file sizes
If ascending is true, the list is sorted from smallest to largest
Otherwise, largest to smallest
cmd line options: -s and -S
*/
func sortSize(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if entry[i].Size > entry[j].Size {
            return !ascending
        }
        if entry[i].Size < entry[j].Size {
            return ascending
        }
        // when multiple files have the same Size, then alphabetize by file name
        return entry[i].Name < entry[j].Name
    })
}

/*
sortModTime sorts the FileStat slice by file modification time
If ascending is true, the list is sorted from oldest to newest
Otherwise, newest to oldest
cmd line options: -d and -D
*/
func sortModTime(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if entry[i].ModTime.After(entry[j].ModTime) {
            return !ascending
        }
        if entry[i].ModTime.Before(entry[j].ModTime) {
            return ascending
        }
        // when multiple files have the same Mod Time, then alphabetize by file name
        return entry[i].Name < entry[j].Name
    })
}

/*
sortName sorts the FileStat slice by file name
If ascending is true, the list is sorted in alphabetical order
Otherwise, reverse alphabetical order
cmd line options: -n and -N
*/
func sortName(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if ascending {
            return entry[i].FullName < entry[j].FullName
        } else {
            return entry[i].FullName > entry[j].FullName
        }
    })
}

/*
sortNameCaseInsensitive sorts the FileStat slice by file name, ignoring case
This is done by making all names lower case before comparing names
If ascending is true, the list is sorted in alphabetical order
Otherwise, reverse alphabetical order
cmd line options: -i and -I
*/
func sortNameCaseInsensitive(entry []FileStat, ascending bool) {
    sort.Slice(entry, func(i, j int) bool {
        if ascending {
            return strings.ToLower(entry[i].FullName) < strings.ToLower(entry[j].FullName)
        } else {
            return strings.ToLower(entry[i].FullName) > strings.ToLower(entry[j].FullName)
        }
    })
}

/*
GetFileInfo will read a list of file names, get the file's timestamp and size,
and create the allEntries slice

Args:
    input: a list of file names, either from a file given on cmd line or read from STDIN

    quiet: when set, errors are not reported to STDERR

Returns:
    a slice of type FileStat containing all files that were successfully examined
*/
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

/*
RenderAllEntries creates a table of all given files which are sorted from the given sort options

Args:
    allEntries: a slice of all files, modification times, sizes, and if the entry is a file, directory, or symbolic link

    addCommas: when set, add a comma as a thousands separator (-c cmd line option)

    convertToMiB: when set, output file size in Mebibytes (-m cmd line option)

    addMilliseconds: when set, output modification times to include thousands of a second (-M cmd line option)

    includeTotals: when set, append a line include summed file sizes and number of files (-t cmd line option)

    onlyFiles: when set, only ouput files and exclude directories, symbolic links (-of cmd line option)

    onlyDirs: when set, only ouput directories and exclude files, symbolic links (-od cmd line option)

    onlyLinks: when set, only ouput symbolic links and exclude files, directories (-ol cmd line option)
*/
func RenderAllEntries(allEntries []FileStat, addCommas bool, convertToMiB bool, addMilliseconds bool, includeTotals bool, onlyFiles bool, onlyDirs bool, onlyLinks bool) {
    var allRows [][]string
    var e FileStat
    var fsize string
    var modtime string
    var totalFileSize int64

    for _,e = range allEntries {
	if onlyFiles && "F" != e.FileType {
		continue
	}
	if onlyDirs && "D" != e.FileType {
		continue
	}
	if onlyLinks && "L" != e.FileType {
		continue
	}
        if includeTotals {
            totalFileSize += e.Size
        }
        if convertToMiB {
            e.Size /= 1048576
        }
        if addCommas {
            fsize = RenderInteger("#,###.",e.Size)
        } else {
            fsize = fmt.Sprintf("%d",e.Size)
        }
        if addMilliseconds {
            modtime = fmt.Sprintf("%s",e.ModTime)[:23]
            if ' ' == modtime[19] {
                modtime = fmt.Sprintf("%s.000", modtime[:19])
            }
        } else {
            modtime = fmt.Sprintf("%s",e.ModTime)[:19]
        }

        allRows = append(allRows, []string{modtime, fsize, fmt.Sprintf("%s",e.FileType), e.FullName})
    }

    if includeTotals {
        tsize := fmt.Sprintf("%d",totalFileSize)
        if convertToMiB {
            totalFileSize /= 1048576
            tsize = fmt.Sprintf("%d",totalFileSize)
        }
        if addCommas {
            tsize = RenderInteger("#,###.",totalFileSize)
        }
        allRows = append(allRows, []string{"", tsize, " ", fmt.Sprintf("(total size for %d files)", len(allRows))})
    }

    table := tablewriter.NewWriter(os.Stdout)
    table.SetAutoWrapText(false)
    table.SetHeader([]string{"Mod Time","Size","Type","Name"})
    table.AppendBulk(allRows)
    if len(allRows) > 0 {
        table.Render()
    }
}

/*
ValidateArgs verify all command line arguments.
It will not allow multiple sort options (such as -s and -d)
It will now allow multiple 'only' options (such as -of and -od)
*/
func ValidateArgs(argsSortSize *bool, argsSortSizeDesc *bool, argsSortModTime *bool, argsSortModTimeDesc *bool, argsSortName *bool, argsSortNameDesc *bool, argsSortNameCaseInsen *bool, argsSortNameCaseInsenDesc *bool, argsOnlyFiles *bool, argsOnlyDirs *bool, argsOnlyLinks *bool ) {
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
        fmt.Fprintf(os.Stderr,"Error: only one 'sort' argument can be given.\n\n")
        os.Exit(2)
    }

    count = 0
    if *argsOnlyFiles { count++ }
    if *argsOnlyDirs { count++ }
    if *argsOnlyLinks { count++ }

    if count > 1 {
        fmt.Fprintf(os.Stderr,"Error: only one 'only' argument can be given.\n\n")
        os.Exit(2)
    }
}

/*
SortAllEntries is used to determine which sorting function to use
At this point, (at most) only one of the *argsSortXXX variables will be true
*/
func SortAllEntries(allEntries []FileStat, argsSortSize *bool, argsSortSizeDesc *bool, argsSortModTime *bool, argsSortModTimeDesc *bool, argsSortName *bool, argsSortNameDesc *bool, argsSortNameCaseInsen *bool, argsSortNameCaseInsenDesc *bool ) {
    if *argsSortSize { sortSize(allEntries,true); return }
    if *argsSortSizeDesc { sortSize(allEntries,false); return }
    if *argsSortModTime { sortModTime(allEntries,true); return }
    if *argsSortModTimeDesc { sortModTime(allEntries,false); return }
    if *argsSortName { sortName(allEntries,true); return }
    if *argsSortNameDesc { sortName(allEntries,false); return }
    if *argsSortNameCaseInsen { sortNameCaseInsensitive(allEntries,true); return }
    if *argsSortNameCaseInsenDesc { sortNameCaseInsensitive(allEntries,false); return }
}

/*
main processes & validates cmd line arguments, reads in file names thus creating allEntries
Next, it sorts the entries and finally renders the results to STDOUT
*/
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
    argsMilliseconds := flag.Bool("M", false, "add milliseconds to file time stamps")
    argsTotals := flag.Bool("t", false, "append total file size and file count")

    argsOnlyFiles := flag.Bool("of", false, "only display files")
    argsOnlyDirs := flag.Bool("od", false, "only display directories")
    argsOnlyLinks := flag.Bool("ol", false, "only display symbolic links")

    flag.Usage = func() {
        pgmName := os.Args[0]
        if(strings.HasPrefix(os.Args[0],"./")) {
            pgmName = os.Args[0][2:]
        }
        fmt.Fprintf(os.Stderr, "\n%s: Get info for a list of files across multiple directories\n", pgmName)
        fmt.Fprintf(os.Stderr, "usage: %s [options] [filename|or blank for STDIN]\n", pgmName)
        fmt.Fprintf(os.Stderr, "       (filename should contain a list of files)\n\n")
        flag.PrintDefaults()
    }

    flag.Parse()
    if *argsVersion {
        fmt.Fprintf(os.Stderr,"version %s\n", version)
        os.Exit(1)
    }

    ValidateArgs(argsSortSize, argsSortSizeDesc, argsSortModTime, argsSortModTimeDesc, argsSortName, argsSortNameDesc, argsSortNameCaseInsen, argsSortNameCaseInsenDesc, argsOnlyFiles, argsOnlyDirs, argsOnlyLinks)
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
    RenderAllEntries(allEntries, *argsCommas, *argsMebibytes, *argsMilliseconds, *argsTotals, *argsOnlyFiles, *argsOnlyDirs, *argsOnlyLinks)
}

