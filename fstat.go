
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
    "encoding/json"
    "fmt"
    "flag"
    "os"
    "path"
    "path/filepath"
    "regexp"
    "sort"
    "strconv"
    "strings"
    "time"

    "github.com/olekukonko/tablewriter"
)

const version = "2.3.1"

type FileStat struct {
    FullName string `json:"fullname"`
    Size int64 `json:"size"`
    ModTime time.Time `json:"modtime"`
    FileType string `json:"filetype"`
}

/*
sortSize sorts the FileStat slice by file sizes
If ascending is true, the list is sorted from smallest to largest
Otherwise, largest to smallest
cmd line options: -ss and -sS
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
        return entry[i].FullName < entry[j].FullName
    })
}

/*
sortModTime sorts the FileStat slice by file modification time
If ascending is true, the list is sorted from oldest to newest
Otherwise, newest to oldest
cmd line options: -sd and -sD
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
        return entry[i].FullName < entry[j].FullName
    })
}

/*
sortName sorts the FileStat slice by file name
If ascending is true, the list is sorted in alphabetical order
Otherwise, reverse alphabetical order
cmd line options: -sn and -sN
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
cmd line options: -si and -sI
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
GetFileList will generate a slice of strings which include all files to be examined

Args:
    input: an input reader that will either be a file given on cmd line or read from STDIN

Returns:
    a slice of strings that are file names
*/
func GetFileList(input *bufio.Scanner) ([]string) {
    var allFilenames []string

    for input.Scan() {
        allFilenames = append(allFilenames, input.Text())
    }
    return allFilenames
}


/*
GetFileInfo will read a list of file names, get the file's timestamp and size,
and create the allEntries slice

Args:
    input: a slice of file names

    quiet: when set, errors are not reported to STDERR (cmd line option: -q)

    excludeDot: when set, exclude dot files (cmd line option: -ed)

    excludeRE: when set, exclude based on this regular expression

    includeRE: when set, only include based on this regular expression

Returns:
    a slice of type FileStat containing all files that were successfully examined
*/
func GetFileInfo(allFilenames []string, quiet bool, excludeDot bool, excludeRE string, includeRE string) ([]FileStat) {
    var allEntries []FileStat
    shouldExcludeRE := false
    shouldIncludeRE := false
    var excludeMatched *regexp.Regexp
    var includeMatched *regexp.Regexp
    var err error

    if len(excludeRE) > 0 {
        excludeMatched, err = regexp.Compile(excludeRE)
        if err != nil {
            fmt.Fprintf(os.Stderr,"Invalid 'exclude' regular expression: %s\n", excludeRE)
            os.Exit(3)
        }
        shouldExcludeRE = true
    }

    if len(includeRE) > 0 {
        includeMatched, err = regexp.Compile(includeRE)
        if err != nil {
            fmt.Fprintf(os.Stderr,"Invalid 'include' regular expression: %s\n", includeRE)
            os.Exit(4)
        }
        shouldIncludeRE = true
    }

    for _,fname:= range(allFilenames) {
        if excludeDot && "." == path.Base(fname)[:1] {
            continue
        }
        if shouldExcludeRE && excludeMatched.Match([]byte(fname)) {
            continue
        }
        if shouldIncludeRE && !includeMatched.Match([]byte(fname)) {
            continue
        }

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

        entry := FileStat{FullName: fname, Size: f.Size(), ModTime: f.ModTime(), FileType: ftype}
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
func RenderAllEntries(allEntries []FileStat, addCommas bool, convertToMiB bool, addMilliseconds bool, includeTotals bool, onlyFiles bool, onlyDirs bool, onlyLinks bool, outputCSV bool, outputHTML bool, outputJSON bool) {
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

    header := []string{"Mod Time","Size","Type","Name"}

    if outputCSV {
        fmt.Printf("\"%s\"\n",strings.Join(header[:],"\",\""))
        for i := 0; i < len(allRows); i++ {
            fmt.Printf("\"%s\"\n",strings.Join(allRows[i][:],"\",\""))
        }
        return
    }

    if outputHTML {
        fmt.Println("<!DOCTYPE html>")
        fmt.Println("<html>")
        fmt.Println("<body>")
        fmt.Println("<table border='1' cellpadding='3' cellspacing='3'>")
        fmt.Printf("<th>%s</th>\n",strings.Join(header[:],"</th><th>"))
        for i := 0; i < len(allRows); i++ {
            fmt.Println("<tr>")
            fmt.Printf("\t<td>%s</td>\n",strings.Join(allRows[i][:],"</td><td>"))
            fmt.Println("</tr>")
        }
        fmt.Println("</table>")
        fmt.Println("</body>")
        fmt.Println("</html>")
        return
    }

    if outputJSON {
        var row FileStat
        var jsonRows []FileStat
        var err error
        layout := "2006-01-02 15:04:05"
        for i := 0; i < len(allRows); i++ {
            row.ModTime, err = time.Parse(layout,allRows[i][0])
            if err != nil {
                fmt.Println(err)
                os.Exit(1)
            }
            row.Size,_ = strconv.ParseInt(strings.Replace(allRows[i][1],",","",-1),10,64)
            row.FileType = allRows[i][2]
            row.FullName = allRows[i][3]
            jsonRows = append(jsonRows,row)
        }
        j, _ := json.MarshalIndent(allRows,"","    ")
        fmt.Println(string(j))

        return
    }

    // by default, output to STDOUT
    if len(allRows) > 0 {
        table := tablewriter.NewWriter(os.Stdout)
        table.SetAutoWrapText(false)
        table.SetHeader(header)
        table.SetAlignment(tablewriter.ALIGN_LEFT)
        table.AppendBulk(allRows)
        table.Render()
    }
}

/*
ValidateArgs verify all command line arguments.
It will not allow multiple sort options (such as -ss and -sd)
It will now allow multiple 'only' options (such as -of and -od)
*/
func ValidateArgs(argsSortSize bool, argsSortSizeDesc bool, argsSortModTime bool, argsSortModTimeDesc bool, argsSortName bool, argsSortNameDesc bool, argsSortNameCaseInsen bool, argsSortNameCaseInsenDesc bool, argsOnlyFiles bool, argsOnlyDirs bool, argsOnlyLinks bool, argsTotals bool, argsOutputCSV bool, argsOutputHTML bool, argsOutputJSON bool ) {
    count := 0
    if argsSortSize { count++ }
    if argsSortSizeDesc { count++ }
    if argsSortModTime { count++ }
    if argsSortModTimeDesc { count++ }
    if argsSortName { count++ }
    if argsSortNameDesc { count++ }
    if argsSortNameCaseInsen { count++ }
    if argsSortNameCaseInsenDesc { count++ }

    if count > 1 {
        fmt.Fprintf(os.Stderr,"Error: only one '-s' sort argument can be given.\n\n")
        os.Exit(2)
    }

    count = 0
    if argsOnlyFiles { count++ }
    if argsOnlyDirs { count++ }
    if argsOnlyLinks { count++ }

    if count > 1 {
        fmt.Fprintf(os.Stderr,"Error: only one '-i' include argument can be given.\n\n")
        os.Exit(2)
    }

    count = 0
    if argsOutputCSV { count++ }
    if argsOutputHTML { count++ }
    if argsOutputJSON { count++ }

    if count > 1 {
        fmt.Fprintf(os.Stderr,"Error: only one '-o' output argument can be given.\n\n")
        os.Exit(2)
    }

    if argsTotals && (argsOutputCSV || argsOutputHTML || argsOutputJSON) {
        fmt.Fprintf(os.Stderr,"Error: -t can not be used with: -oc, -oh, or -oj\n\n")
        os.Exit(2)
    }
}

/*
SortAllEntries is used to determine which sorting function to use
At this point, (at most) only one of the *argsSortXXX variables will be true
*/
func SortAllEntries(allEntries []FileStat, argsSortSize bool, argsSortSizeDesc bool, argsSortModTime bool, argsSortModTimeDesc bool, argsSortName bool, argsSortNameDesc bool, argsSortNameCaseInsen bool, argsSortNameCaseInsenDesc bool ) {
    if argsSortSize { sortSize(allEntries,true); return }
    if argsSortSizeDesc { sortSize(allEntries,false); return }
    if argsSortModTime { sortModTime(allEntries,true); return }
    if argsSortModTimeDesc { sortModTime(allEntries,false); return }
    if argsSortName { sortName(allEntries,true); return }
    if argsSortNameDesc { sortName(allEntries,false); return }
    if argsSortNameCaseInsen { sortNameCaseInsensitive(allEntries,true); return }
    if argsSortNameCaseInsenDesc { sortNameCaseInsensitive(allEntries,false); return }
}

/*
main processes & validates cmd line arguments, reads in file names thus creating allEntries
Next, it sorts the entries and finally renders the results to STDOUT
*/
func main() {
    argsSortSize := flag.Bool("ss", false, "sort by file size")
    argsSortSizeDesc := flag.Bool("sS", false, "sort by file size, descending")

    argsSortModTime := flag.Bool("sd", false, "sort by file modified date")
    argsSortModTimeDesc := flag.Bool("sD", false, "sort by file modified date, newest first")

    argsSortName := flag.Bool("sn", false, "sort by file name")
    argsSortNameDesc := flag.Bool("sN", false, "sort by file name, reverse alphabetical order")

    argsSortNameCaseInsen := flag.Bool("si", false, "sort by file name, ignore case")
    argsSortNameCaseInsenDesc := flag.Bool("sI", false, "sort by file name, ignore case, reverse alphabetical order")

    argsVersion := flag.Bool("v", false, "show program version and then exit")
    argsQuiet := flag.Bool("q", false, "do not display file errors")
    argsCommas := flag.Bool("c", false, "add comma thousands separator to file sizes")
    argsMebibytes := flag.Bool("m", false, "convert file sizes to mebibytes")
    argsMilliseconds := flag.Bool("M", false, "add milliseconds to file time stamps")
    argsTotals := flag.Bool("t", false, "append total file size and file count")

    argsOnlyFiles := flag.Bool("if", false, "include only files")
    argsOnlyDirs := flag.Bool("id", false, "include only directories")
    argsOnlyLinks := flag.Bool("il", false, "include only symbolic links")

    argsOutputCSV := flag.Bool("oc", false, "ouput to CSV format")
    argsOutputHTML := flag.Bool("oh", false, "ouput to HTML format")
    argsOutputJSON := flag.Bool("oj", false, "ouput to JSON format")

    argsFilenames := flag.String("f", "", "use these files instead of from a file or STDIN, can include wildcards")
    argsExcludeDot := flag.Bool("ed", false, "exclude-dot, exclude anything starting with a dot")
    argsExcludeRE := flag.String("er", "", "exclude-regexp, exclude based on given regular expression; use .* instead of just *")
    argsIncludeRE := flag.String("ir", "", "include-regexp, only include based on given regular expression; use .* instead of just *")

    flag.Usage = func() {
        pgmName := os.Args[0]
        if(strings.HasPrefix(os.Args[0],"./")) {
            pgmName = os.Args[0][2:]
        }
        fmt.Fprintf(os.Stderr, "\n%s: Get info for a list of files across multiple directories\n", pgmName)
        fmt.Fprintf(os.Stderr, "usage: %s [options] [filename|or blank for STDIN]\n", pgmName)
        fmt.Fprintf(os.Stderr, "       (this file should contain a list of files to process)\n\n")
        flag.PrintDefaults()
        fmt.Fprintf(os.Stderr, "\nNote: -er precedes -ir\n\n")
    }

    flag.Parse()
    if *argsVersion {
        fmt.Fprintf(os.Stderr,"version %s\n", version)
        os.Exit(1)
    }

    ValidateArgs(*argsSortSize, *argsSortSizeDesc, *argsSortModTime, *argsSortModTimeDesc, *argsSortName, *argsSortNameDesc, *argsSortNameCaseInsen, *argsSortNameCaseInsenDesc, *argsOnlyFiles, *argsOnlyDirs, *argsOnlyLinks, *argsTotals, *argsOutputCSV, *argsOutputHTML, *argsOutputJSON)
    args := flag.Args()
    var allFilenames []string

    // get a list of filenames by either using -f
    // or by reading from a file
    // or by reading from STDIN
    if len(*argsFilenames) > 0 { // using -f
        // -f can be a space delimited list of filename wildcards (aka Globs)
        // iterate through all of these globs to create a unique list of files named allFilenames
        // (this is done by using a temporary map named allGlobbedNames
        var allGlobs []string
        var n,m int
        allGlobbedNames := make(map[string]int)

        // get slice of wildcards
        fileglobs := strings.Fields(*argsFilenames)
        for n=0; n < len(fileglobs); n++ {
            allGlobs = append(allGlobs,fileglobs[n])
        }

        // create a slice of files in one of those wildcard entries named currentFilelist
        for n=0; n < len(allGlobs); n++ {
            currentFilelist, err := filepath.Glob(allGlobs[n])
            if err != nil {
                fmt.Fprintf(os.Stderr, "%s\n", err)
                continue
            }
            // add all of these file names to a 'global' map of files
            // duplicates file names are discarded
            for m=0; m < len(currentFilelist); m++ {
                allGlobbedNames[currentFilelist[m]] = 0
            }
        }
        // from the allGlobbedNames map, create the allFilenames slice
        // (which is the file goal)
        for key,_ := range allGlobbedNames {
            allFilenames = append(allFilenames,key)
        }
    } else { // using a filename or STDIN
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
        allFilenames = GetFileList(input)
    }

    allEntries := GetFileInfo(allFilenames, *argsQuiet, *argsExcludeDot, *argsExcludeRE, *argsIncludeRE)
    SortAllEntries(allEntries, *argsSortSize, *argsSortSizeDesc, *argsSortModTime, *argsSortModTimeDesc, *argsSortName, *argsSortNameDesc, *argsSortNameCaseInsen, *argsSortNameCaseInsenDesc)
    RenderAllEntries(allEntries, *argsCommas, *argsMebibytes, *argsMilliseconds, *argsTotals, *argsOnlyFiles, *argsOnlyDirs, *argsOnlyLinks, *argsOutputCSV, *argsOutputHTML, *argsOutputJSON)
}

