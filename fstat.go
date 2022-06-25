/*

fstat.go
-John Taylor
Mar 2019

Get info for a list of files across multiple directories

To compile:
go build -ldflags="-s -w" fstat.go render_number.go

MIT License; Copyright (c) 2021 John Taylor
Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

*/

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jftuga/ellipsis"
	"github.com/jftuga/termsize"
	"github.com/olekukonko/tablewriter"
)

const version = "2.6.11"
const minTermWidth = 49

// used for -do and -dn cmd line options
const (
	wantOlder = iota
	wantNewer
)
const dateFormat = "20060102"

// FileStat - metadata for each entry
type FileStat struct {
	FullName string    `json:"fullname"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"modtime"`
	FileType string    `json:"filetype"`
}

// shortenFileName - shorten file names in the last column
// this is done by inserting "..." in the middle of a long file path
func shortenFileName(allRows [][]string, maxWidth int) [][]string {
	newRows := make([][]string, len(allRows))

	for i := 0; i < len(allRows); i++ {
		row := allRows[i]
		row[3] = ellipsis.Shorten(row[3], maxWidth)
		//fmt.Println(row)
		newRows = append(newRows, row)
	}
	return newRows
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
		}
		return entry[i].FullName > entry[j].FullName
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
		}
		return strings.ToLower(entry[i].FullName) > strings.ToLower(entry[j].FullName)
	})
}

/*
GetFileList will generate a slice of strings which include all files to be examined

Args:
    input: an input reader that will either be a file given on cmd line or read from STDIN

Returns:
    a slice of strings that are file names
*/
func GetFileList(input *bufio.Scanner) []string {
	var allFilenames []string

	for input.Scan() {
		allFilenames = append(allFilenames, input.Text())
	}
	return allFilenames
}

/*
roundToLocalTime converts from UTC to Local time and rounds the day
to either the start of the day or end of the day depending on the value of olderOrNewer

Args:
    olderOrNewer: should be either wantOlder or wantNewer, depending on which files you want

    modTime: the time in YYYYMMDD format

Returns:
    a rounded time, in the current Local time zone

    if wantOlder, modTime is rounded up to the last nanosecond of the given day
    Example: given modTime of 20190325; then "2019-03-25 23:59:59.999999999 -0400 EDT" is returned
    (when Local time zone is: Eastern Daylight Savings)

    otherwise (wantNewer), modTime is rounded down to 1 nanosecond before the day starts
    Example: given modTime of 20190325; then "2019-03-24 23:59:59.999999999 -0400 EDT" is returned
    (when Local time zone is: Eastern Daylight Savings)
*/
//goland:noinspection GoUnhandledErrorResult
func roundToLocalTime(olderOrNewer int, modTime string) time.Time {
	// set up time.Time variables for dateOlder and dateNewer; -do and -dn
	// roundedModTime will be rounded down
	roundedModTime, err := time.Parse(dateFormat, modTime)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error when parsing date:", modTime)
		fmt.Fprintln(os.Stderr, "Date format should be  : YYYYMMDD")
		os.Exit(5)
	}

	//fmt.Println("[start.1] roundedModTime: ", roundedModTime)
	roundedModTime = roundedModTime.In(time.Local)
	//fmt.Println("[start.2] roundedModTime: ", roundedModTime)
	roundedModTime = time.Date(roundedModTime.Year(), roundedModTime.Month(), roundedModTime.Day(), 23, 59, 59, 999999999, roundedModTime.Location())
	//fmt.Println("[start.3] roundedModTime: ", roundedModTime)

	if olderOrNewer == wantOlder {
		// add 1 day so that a file with a timestamp of 23:59:59 (of the same day) will be included
		roundedModTime = roundedModTime.Add(time.Hour * 24)
		//fmt.Println("[older.4] roundedModTime: ", roundedModTime)
	}
	return roundedModTime
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

    dateNewer: when set, only include if date is equal or newer that the given YYYYMMDD formatted date

    dateOlder: when set, only include if date is equal or older that the given YYYYMMDD formatted date

    sizeSmaller: when set, only include if file size is equal or smaller that given value (in bytes)

    sizeLarger: when set, only include if file size is equal or larger that given value (in bytes)

Returns:
    a slice of type FileStat containing all files that were successfully examined
*/
func GetFileInfo(allFilenames []string, quiet bool, excludeDot bool, excludeRE string, includeRE string, dateNewer string, dateOlder string, sizeSmaller int64, sizeLarger int64) []FileStat {
	var allEntries []FileStat
	shouldExcludeRE := false
	shouldIncludeRE := false
	var excludeMatched *regexp.Regexp
	var includeMatched *regexp.Regexp
	var err error

	if len(excludeRE) > 0 {
		excludeMatched, err = regexp.Compile(excludeRE)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid 'exclude' regular expression: %s\n", excludeRE)
			os.Exit(3)
		}
		shouldExcludeRE = true
	}

	if len(includeRE) > 0 {
		includeMatched, err = regexp.Compile(includeRE)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid 'include' regular expression: %s\n", includeRE)
			os.Exit(4)
		}
		shouldIncludeRE = true
	}

	// set up time.Time variables for dateOlder and dateNewer; -do and -dn
	var olderModTime, newerModTime time.Time
	useOlder := false
	useNewer := false
	if len(dateNewer) > 0 {
		useNewer = true
		newerModTime = roundToLocalTime(wantNewer, dateNewer)
	}
	if len(dateOlder) > 0 {
		useOlder = true
		olderModTime = roundToLocalTime(wantOlder, dateOlder)
	}

	// iterate through each file and get its os.Lstat()
	pathSepDot := fmt.Sprintf("%c.", os.PathSeparator)
	for _, fname := range allFilenames {
		// check excludeDot; -ed
		if excludeDot && ("." == path.Base(fname)[:1] || strings.Contains(fname, pathSepDot)) {
			continue
		}

		// check excludeRE and includeRE; -er and -ir
		if shouldExcludeRE && excludeMatched.Match([]byte(fname)) {
			continue
		}
		if shouldIncludeRE && !includeMatched.Match([]byte(fname)) {
			continue
		}

		f, err := os.Lstat(fname)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			}
			continue
		}

		// check dateOlder and dateNewer; -do and -dn
		if useOlder && f.ModTime().After(olderModTime) {
			continue
		}
		if useNewer && f.ModTime().Before(newerModTime) {
			continue
		}

		var ftype = "?"
		if f.Mode().IsRegular() {
			ftype = "F"
		} else if f.IsDir() {
			ftype = "D"
		} else if f.Mode()&os.ModeSymlink == os.ModeSymlink {
			ftype = "L"
		}

		// check file sizes; -szs and -szl
		if sizeSmaller > 0 && f.Size() > sizeSmaller && "F" == ftype {
			continue
		}
		// check file sizes; -szs and -szl
		if sizeLarger > 0 && f.Size() < sizeLarger && "F" == ftype {
			continue
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

    onlyFiles: when set, only output files and exclude directories, symbolic links (-of cmd line option)

    onlyDirs: when set, only output directories and exclude files, symbolic links (-od cmd line option)

    onlyLinks: when set, only output symbolic links and exclude files, directories (-ol cmd line option)

	longFileNames: when set, do not use ellipses to shorten file names (-long cmd line option)

	longWidth: when set, use this at the max line width (-longwidth cmd line option)

*/
func RenderAllEntries(allEntries []FileStat, addCommas bool, convertToMiB bool, addMilliseconds bool, includeTotals bool, onlyFiles bool, onlyDirs bool, onlyLinks bool, outputCSV bool, outputHTML bool, outputJSON bool, longFileNames bool, longWidth int) {
	var allRows [][]string
	var e FileStat
	var fsize string
	var modtime string
	var totalFileSize int64
	var totalFileCount int64
	var totalDirCount int64
	var totalSymLinkCount int64

	for _, e = range allEntries {
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
			if "F" == e.FileType {
				totalFileSize += e.Size
				totalFileCount++
			}
			if "D" == e.FileType {
				totalDirCount++
			}
			if "L" == e.FileType {
				totalSymLinkCount++
			}
		}
		if convertToMiB {
			e.Size /= 1048576
		}
		if addCommas {
			fsize = RenderInteger("#,###.", e.Size)
		} else {
			fsize = fmt.Sprintf("%d", e.Size)
		}
		if addMilliseconds {
			modtime = fmt.Sprintf("%s", e.ModTime)[:23]
			if ' ' == modtime[19] {
				modtime = fmt.Sprintf("%s.000", modtime[:19])
			}
		} else {
			modtime = fmt.Sprintf("%s", e.ModTime)[:19]
		}

		allRows = append(allRows, []string{modtime, fsize, fmt.Sprintf("%s", e.FileType), e.FullName})
	}

	if includeTotals {
		tsize := fmt.Sprintf("%d", totalFileSize)
		if convertToMiB {
			totalFileSize /= 1048576
			tsize = fmt.Sprintf("%d", totalFileSize)
		}
		if addCommas {
			tsize = RenderInteger("#,###.", totalFileSize)
		}
		allRows = append(allRows, []string{"", tsize, " ", fmt.Sprintf("  (total size for %d files)", totalFileCount)})

		var averageFileSize float64
		if totalFileCount > 0 {
			averageFileSize = float64(totalFileSize / totalFileCount)
		}

		var averageFilesPerDir float64
		if totalFileCount > 0 && totalDirCount > 0 {
			averageFilesPerDir = float64(totalFileCount / totalDirCount)
		}

		asize := fmt.Sprintf("%.0f", averageFileSize)
		dsize := fmt.Sprintf("%.0f", averageFilesPerDir)
		if addCommas {
			asize = RenderFloat("#,###.", averageFileSize)
			dsize = RenderFloat("#,###.", averageFilesPerDir)
		}
		if len(allRows) > 0 {
			allRows = append(allRows, []string{"", asize, " ", fmt.Sprintf("(average size for %d files)", totalFileCount)})
		}
		if totalDirCount > 0 {
			allRows = append(allRows, []string{"", fmt.Sprintf("%d", totalDirCount), " ", "(num of directories)"})
		}
		if averageFilesPerDir > 0 {
			allRows = append(allRows, []string{"", dsize, " ", "(average num of files per directory)"})
		}
		if totalSymLinkCount > 0 {
			allRows = append(allRows, []string{"", fmt.Sprintf("%d", totalSymLinkCount), " ", "(num of sym links)"})
		}
	}

	header := []string{"Mod Time", "Size", "Type", "Name"}

	if outputCSV {
		fmt.Printf("\"%s\"\n", strings.Join(header[:], "\",\""))
		for i := 0; i < len(allRows); i++ {
			fmt.Printf("\"%s\"\n", strings.Join(allRows[i][:], "\",\""))
		}
		return
	}

	if outputHTML {
		fmt.Println("<!DOCTYPE html>")
		fmt.Println("<html>")
		fmt.Println("<body>")
		fmt.Println("<table border='1' cellpadding='3' cellspacing='3'>")
		fmt.Printf("<th>%s</th>\n", strings.Join(header[:], "</th><th>"))
		for i := 0; i < len(allRows); i++ {
			fmt.Println("<tr>")
			fmt.Printf("\t<td>%s</td>\n", strings.Join(allRows[i][:], "</td><td>"))
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
			row.ModTime, err = time.Parse(layout, allRows[i][0])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			row.Size, _ = strconv.ParseInt(strings.Replace(allRows[i][1], ",", "", -1), 10, 64)
			row.FileType = allRows[i][2]
			row.FullName = allRows[i][3]
			jsonRows = append(jsonRows, row)
		}
		j, _ := json.MarshalIndent(allRows, "", "    ")
		fmt.Println(string(j))

		return
	}

	// by default, output to STDOUT
	if len(allRows) > 0 {
		maxWidth := 3000
		if longFileNames == false {
			maxWidth = termsize.Width() - minTermWidth
			if longWidth > 0 {
				maxWidth = longWidth - minTermWidth + 2
			}
			if maxWidth < minTermWidth {
				maxWidth = minTermWidth
			}
		}

		allRows = shortenFileName(allRows, maxWidth)
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetHeader(header)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		table.AppendBulk(allRows)
		table.Render()
	}
}

/*
ValidateArgs verify all command line arguments.
It will not allow multiple sort options (such as -ss and -sd)
It will now allow multiple 'only' options (such as -of and -od)
*/
func ValidateArgs(argsSortSize bool, argsSortSizeDesc bool, argsSortModTime bool, argsSortModTimeDesc bool, argsSortName bool, argsSortNameDesc bool, argsSortNameCaseInsen bool, argsSortNameCaseInsenDesc bool, argsOnlyFiles bool, argsOnlyDirs bool, argsOnlyLinks bool, argsTotals bool, argsOutputCSV bool, argsOutputHTML bool, argsOutputJSON bool, dateOlder string, dateNewer string, sizeSmaller int64, sizeLarger int64, longFileNames bool, longWidth int) {
	count := 0
	if argsSortSize {
		count++
	}
	if argsSortSizeDesc {
		count++
	}
	if argsSortModTime {
		count++
	}
	if argsSortModTimeDesc {
		count++
	}
	if argsSortName {
		count++
	}
	if argsSortNameDesc {
		count++
	}
	if argsSortNameCaseInsen {
		count++
	}
	if argsSortNameCaseInsenDesc {
		count++
	}

	if count > 1 {
		fmt.Fprintf(os.Stderr, "Error: only one '-s' sort argument can be given.\n\n")
		os.Exit(2)
	}

	count = 0
	if argsOnlyFiles {
		count++
	}
	if argsOnlyDirs {
		count++
	}
	if argsOnlyLinks {
		count++
	}

	if count > 1 {
		fmt.Fprintf(os.Stderr, "Error: only one '-i' include argument can be given.\n\n")
		os.Exit(2)
	}

	count = 0
	if argsOutputCSV {
		count++
	}
	if argsOutputHTML {
		count++
	}
	if argsOutputJSON {
		count++
	}

	if count > 1 {
		fmt.Fprintf(os.Stderr, "Error: only one '-o' output argument can be given.\n\n")
		os.Exit(2)
	}

	if argsTotals && (argsOutputCSV || argsOutputHTML || argsOutputJSON) {
		fmt.Fprintf(os.Stderr, "Error: -t can not be used with: -oc, -oh, or -oj\n\n")
		os.Exit(2)
	}

	// make sure dateNewer is not newer than dateOlder
	var older, newer time.Time
	var err error
	if len(dateOlder) > 0 && len(dateNewer) > 0 {
		older, err = time.Parse(dateFormat, dateOlder)
		if err != nil {
			//goland:noinspection GoUnhandledErrorResult
			fmt.Fprintln(os.Stderr, "Error when parsing date for '-do':", dateOlder)
			os.Exit(2)
		}
		newer, err = time.Parse(dateFormat, dateNewer)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error when parsing date for '-dn':", dateNewer)
			os.Exit(2)
		}
		if older.After(newer) {
			//goland:noinspection GoUnhandledErrorResult
			fmt.Fprintln(os.Stderr, "Error: '-dn' date is newer than '-do'")
			os.Exit(2)
		}
	}

	// make sure sizeSmaller is not smaller than sizeLarger
	if sizeSmaller > 0 && sizeSmaller < sizeLarger {
		fmt.Fprintln(os.Stderr, "Error: '-szs' file size is smaller than '-szl'")
		os.Exit(2)
	}

	// these are mutually exclusive
	if longFileNames == true && longWidth > 0 {
		fmt.Fprintln(os.Stderr, "Error: '-long' and '-longwidth' are mutually exclusive")
		os.Exit(2)
	}
}

/*
SortAllEntries is used to determine which sorting function to use
At this point, (at most) only one of the *argsSortXXX variables will be true
*/
func SortAllEntries(allEntries []FileStat, argsSortSize bool, argsSortSizeDesc bool, argsSortModTime bool, argsSortModTimeDesc bool, argsSortName bool, argsSortNameDesc bool, argsSortNameCaseInsen bool, argsSortNameCaseInsenDesc bool) {
	if argsSortSize {
		sortSize(allEntries, true)
		return
	}
	if argsSortSizeDesc {
		sortSize(allEntries, false)
		return
	}
	if argsSortModTime {
		sortModTime(allEntries, true)
		return
	}
	if argsSortModTimeDesc {
		sortModTime(allEntries, false)
		return
	}
	if argsSortName {
		sortName(allEntries, true)
		return
	}
	if argsSortNameDesc {
		sortName(allEntries, false)
		return
	}
	if argsSortNameCaseInsen {
		sortNameCaseInsensitive(allEntries, true)
		return
	}
	if argsSortNameCaseInsenDesc {
		sortNameCaseInsensitive(allEntries, false)
		return
	}
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

	argsOutputCSV := flag.Bool("oc", false, "output to CSV format")
	argsOutputHTML := flag.Bool("oh", false, "output to HTML format")
	argsOutputJSON := flag.Bool("oj", false, "output to JSON format")

	argsFilenames := flag.String("f", "", "use these files instead of from a file or STDIN, can include wildcards")
	argsExcludeDot := flag.Bool("ed", false, "exclude-dot, exclude all dot files and directories")
	argsExcludeRE := flag.String("er", "", "exclude-regexp, exclude based on given regular expression; use .* instead of just *")
	argsIncludeRE := flag.String("ir", "", "include-regexp, only include based on given regular expression; use .* instead of just *")

	argsDateNewer := flag.String("dn", "", "only include if date is equal or newer than given YYYYMMDD date")
	argsDateOlder := flag.String("do", "", "only include if date is equal or older than given YYYYMMDD date")

	argsSizeSmaller := flag.Int64("szs", 0, "only include if file size is equal or smaller than given value (in bytes)")
	argsSizeLarger := flag.Int64("szl", 0, "only include if file size is equal or larger than given value (in bytes)")

	argsLongFileNames := flag.Bool("long", false, "Don't use ellipses for long file names; useful when piping or using redirection")
	argsLongWidth := flag.Int("longwidth", 0, "Set max width; Useful when piping or using redirection")

	flag.Usage = func() {
		pgmName := os.Args[0]
		if strings.HasPrefix(os.Args[0], "./") {
			pgmName = os.Args[0][2:]
		}
		fmt.Fprintf(os.Stderr, "\n%s: Get info for a list of files across multiple directories\n", pgmName)
		fmt.Fprintf(os.Stderr, "usage: %s [options] [filename|or blank for STDIN]\n", pgmName)
		fmt.Fprintf(os.Stderr, "       (this file should contain a list of files to process)\n\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nNotes:\n")
		fmt.Fprintf(os.Stderr, "  (1) -er precedes -ir\n")
		fmt.Fprintf(os.Stderr, "  (2) Use '(?i)' at the beginning of a regex to make it case insensitive\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()
	if *argsVersion {
		fmt.Fprintf(os.Stderr, "version %s\n", version)
		os.Exit(1)
	}

	ValidateArgs(*argsSortSize, *argsSortSizeDesc, *argsSortModTime, *argsSortModTimeDesc, *argsSortName, *argsSortNameDesc, *argsSortNameCaseInsen, *argsSortNameCaseInsenDesc, *argsOnlyFiles, *argsOnlyDirs, *argsOnlyLinks, *argsTotals, *argsOutputCSV, *argsOutputHTML, *argsOutputJSON, *argsDateNewer, *argsDateOlder, *argsSizeSmaller, *argsSizeLarger, *argsLongFileNames, *argsLongWidth)
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
		var n, m int
		allGlobbedNames := make(map[string]int)

		// get slice of wildcards
		fileglobs := strings.Fields(*argsFilenames)
		for n = 0; n < len(fileglobs); n++ {
			allGlobs = append(allGlobs, fileglobs[n])
		}

		// create a slice of files in one of those wildcard entries named currentFilelist
		for n = 0; n < len(allGlobs); n++ {
			currentFilelist, err := filepath.Glob(allGlobs[n])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				continue
			}
			// add all of these file names to a 'global' map of files
			// duplicates file names are discarded
			for m = 0; m < len(currentFilelist); m++ {
				allGlobbedNames[currentFilelist[m]] = 0
			}
		}
		// from the allGlobbedNames map, create the allFilenames slice
		// (which is the file goal)
		for key := range allGlobbedNames {
			allFilenames = append(allFilenames, key)
		}
		if len(allFilenames) == 0 {
			fmt.Fprintf(os.Stderr, "Error: -f did not match any file names.\n\n")
			os.Exit(3)
		}
		if len(allFilenames) == 1 {
			fmt.Fprintf(os.Stderr, "Warning: -f only matched one file name.\n\n")
		}
	} else { // using a filename or STDIN
		var input *bufio.Scanner
		usingFile := ""
		if 0 == len(args) { // read from STDIN
			input = bufio.NewScanner(os.Stdin)
			usingFile = "STDIN"
		} else { // read from filename
			fname := args[0]
			usingFile = fname
			file, err := os.Open(fname)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			defer file.Close()
			input = bufio.NewScanner(file)
		}
		allFilenames = GetFileList(input)
		if len(allFilenames) == 0 {
			fmt.Fprintf(os.Stderr, "Error: No files were listed in '%s'\n\n", usingFile)
			os.Exit(3)
		}
	}

	allEntries := GetFileInfo(allFilenames, *argsQuiet, *argsExcludeDot, *argsExcludeRE, *argsIncludeRE, *argsDateNewer, *argsDateOlder, *argsSizeSmaller, *argsSizeLarger)
	SortAllEntries(allEntries, *argsSortSize, *argsSortSizeDesc, *argsSortModTime, *argsSortModTimeDesc, *argsSortName, *argsSortNameDesc, *argsSortNameCaseInsen, *argsSortNameCaseInsenDesc)
	RenderAllEntries(allEntries, *argsCommas, *argsMebibytes, *argsMilliseconds, *argsTotals, *argsOnlyFiles, *argsOnlyDirs, *argsOnlyLinks, *argsOutputCSV, *argsOutputHTML, *argsOutputJSON, *argsLongFileNames, *argsLongWidth)
}
