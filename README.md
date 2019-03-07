# fstat

`fstat` is useful when you need to obtain file names, sizes, and timestamps across multiple directories.  You can also sort by timestamp, file size, and file name (both case-sensitive and case-insensitive). 

The [Releases Page](https://github.com/jftuga/fstat/releases) contains binaries for Windows, MacOS, Linux and FreeBSD.

For the `TYPE` column (see examples below):

* `F` represents regular file
* `D` represents directory
* `L` represents symbolic link

___

### Usage
```
fstat: Get info for a list of files across multiple directories

usage: fstat [options] [filename|or blank for STDIN]
       filename should contain a list of files

Usage of fstat:
  -D	sort by file modified date, newest first
  -I	case-insensitive sort by file name, reverse alphabetical order
  -M	add milliseconds to file time stamps
  -N	sort by file name, reverse alphabetical order
  -S	sort by file size, descending
  -c	add comma thousands separator to file sizes
  -d	sort by file modified date
  -i	case-insensitive sort by file name
  -m	convert file sizes to mebibytes
  -n	sort by file name
  -q	do not display file errors
  -s	sort by file size
  -t	append total file size and file count
  -v	show program version and then exit
```

___

### Examples

Running `fstat` on Windows with no options:
```
c:\> dir /s/b "c:\Program Files\Microsoft Office\*.exe" | fstat.exe

+---------------------+---------+------+---------------------------------------------------------------------------------------------------+
|      MOD TIME       |  SIZE   | TYPE |                                                           NAME                                    |
+---------------------+---------+------+---------------------------------------------------------------------------------------------------+
| 2019-02-20 14:35:11 |  414360 | F    | c:\Program Files\Microsoft Office\root\Office16\VPREVIEW.EXE                                      |
| 2019-02-20 14:35:11 | 1966392 | F    | c:\Program Files\Microsoft Office\root\Office16\WINWORD.EXE                                       |
| 2019-02-20 14:35:11 |   36936 | F    | c:\Program Files\Microsoft Office\root\Office16\Wordconv.exe                                      |
| 2018-12-05 10:12:30 | 3026088 | F    | c:\Program Files\Microsoft Office\root\Office16\WORDICON.EXE                                      |
| 2018-12-05 10:12:31 | 3696296 | F    | c:\Program Files\Microsoft Office\root\Office16\XLICONS.EXE                                       |
| 2018-12-05 10:10:48 |  289584 | F    | c:\Program Files\Microsoft Office\root\Office16\1033\VISEVMON.EXE                                 |
| 2019-02-20 14:34:56 |   40264 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\Common.DBConnection.exe                       |
| 2019-02-20 14:34:56 |   39032 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\Common.DBConnection64.exe                     |
| 2018-12-05 10:11:38 |   33592 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\Common.ShowHelp.exe                           |
| 2019-02-20 14:34:56 |  186704 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\DATABASECOMPARE.EXE                           |
| 2018-12-05 10:12:18 |  267384 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\filecompare.exe                               |
| 2019-02-20 14:34:56 |  465528 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\SPREADSHEETCOMPARE.EXE                        |
| 2018-12-05 10:11:37 |   82240 | F    | c:\Program Files\Microsoft Office\root\Office16\SkypeSrv\SKYPESERVER.EXE                          |
| 2019-01-10 10:06:19 |  372864 | F    | c:\Program Files\Microsoft Office\root\vfs\ProgramFilesX64\Microsoft Office\Office16\MSOHTMED.EXE |
+---------------------+---------+------+---------------------------------------------------------------------------------------------------+
```

Running `fstat` on Linux, using `-s` to sort by file size
```
user@debian:~$ find /usr/share -name '*exec*' | fstat -s

+---------------------+-------+------+-----------------------------------------------------+
|      MOD TIME       | SIZE  | TYPE |                        NAME                         |
+---------------------+-------+------+-----------------------------------------------------+
| 2016-02-19 03:25:10 |    10 | L    | /usr/share/terminfo/e/exec80                        |
| 2019-02-13 18:30:49 |    25 | L    | /usr/share/man/man8/systemd-kexec.service.8.gz      |
| 2016-02-19 03:22:31 |  1081 | F    | /usr/share/terminfo/o/osexec                        |
| 2018-04-09 07:47:32 |  1746 | F    | /usr/share/man/man8/pam_exec.8.gz                   |
| 2018-09-27 18:09:42 |  2690 | F    | /usr/share/man/man8/aa-exec.8.gz                    |
| 2018-11-28 19:19:27 |  2699 | F    | /usr/share/mime/application/x-executable.xml        |
| 2018-11-28 19:19:28 |  2865 | F    | /usr/share/mime/application/x-pef-executable.xml    |
| 2019-01-15 08:52:42 |  3440 | F    | /usr/share/man/man1/pkexec.1.gz                     |
| 2018-11-28 19:19:28 |  3491 | F    | /usr/share/mime/application/x-ms-dos-executable.xml |
| 2016-11-24 15:50:23 |  3910 | F    | /usr/share/vim/vim74/syntax/focexec.vim             |
| 2019-02-13 18:30:43 | 12619 | F    | /usr/share/man/man5/systemd.exec.5.gz               |
+---------------------+-------+------+-----------------------------------------------------+
```

Running `fstat` on MacOS, using `-S -c` to sort by file size decending, adding commas to file size
```
macbook:fstat user$ find /Applications/Safari.app/Contents/ -name G\*nib|./fstat -S -c
+---------------------+--------+------+--------------------------------------------------------------------------------+
|      MOD TIME       |  SIZE  | TYPE |                                      NAME                                      |
+---------------------+--------+------+--------------------------------------------------------------------------------+
| 2019-01-14 21:29:46 | 34,759 | F    | /Applications/Safari.app/Contents//Resources/Base.lproj/GeneralPreferences.nib |
| 2019-01-14 21:44:24 | 31,084 | F    | /Applications/Safari.app/Contents//Resources/ko.lproj/GeneralPreferences.nib   |
+---------------------+--------+------+--------------------------------------------------------------------------------+
```

Running `fstat` on Linux, using `-D` to sort by modification time, newest timestamp first
```
user@debian:~$ find /lib | grep cryptsetup | ./fstat -D

+---------------------+--------+------+-------------------------------------------------------------+
|      MOD TIME       |  SIZE  | TYPE |                            NAME                             |
+---------------------+--------+------+-------------------------------------------------------------+
| 2019-02-13 18:31:00 |  72296 | F    | /lib/systemd/system-generators/systemd-cryptsetup-generator |
| 2019-02-13 18:30:59 |  92752 | F    | /lib/systemd/systemd-cryptsetup                             |
| 2019-02-13 18:30:47 |     20 | L    | /lib/systemd/system/sysinit.target.wants/cryptsetup.target  |
| 2019-02-13 18:30:36 |    394 | F    | /lib/systemd/system/cryptsetup-pre.target                   |
| 2019-02-13 18:30:36 |    366 | F    | /lib/systemd/system/cryptsetup.target                       |
| 2018-03-26 12:32:43 |   4096 | D    | /lib/cryptsetup/checks                                      |
| 2018-03-26 12:32:43 |   4096 | D    | /lib/cryptsetup                                             |
| 2018-03-26 12:32:43 |   4096 | D    | /lib/cryptsetup/scripts                                     |
| 2018-03-26 12:31:16 |     22 | L    | /lib/x86_64-linux-gnu/libcryptsetup.so.4                    |
| 2017-09-06 06:08:21 |  14928 | F    | /lib/cryptsetup/askpass                                     |
| 2017-09-06 06:08:21 | 158920 | F    | /lib/x86_64-linux-gnu/libcryptsetup.so.4.6.0                |
| 2017-09-06 06:08:21 |  10552 | F    | /lib/cryptsetup/scripts/passdev                             |
| 2017-09-06 06:08:16 |   1040 | F    | /lib/cryptsetup/checks/blkid                                |
| 2017-09-06 06:08:16 |  19047 | F    | /lib/cryptsetup/cryptdisks.functions                        |
| 2017-09-06 06:08:16 |   1214 | F    | /lib/cryptsetup/scripts/decrypt_derived                     |
| 2017-09-06 06:08:16 |    576 | F    | /lib/cryptsetup/scripts/decrypt_gnupg                       |
| 2017-09-06 06:08:16 |   3042 | F    | /lib/cryptsetup/scripts/decrypt_keyctl                      |
| 2017-09-06 06:08:16 |   1724 | F    | /lib/cryptsetup/scripts/decrypt_openct                      |
| 2017-09-06 06:08:16 |   1414 | F    | /lib/cryptsetup/scripts/decrypt_opensc                      |
| 2017-09-06 06:08:16 |    347 | F    | /lib/cryptsetup/scripts/decrypt_ssl                         |
| 2017-09-06 06:08:16 |    387 | F    | /lib/cryptsetup/checks/ext2                                 |
| 2017-09-06 06:08:16 |    148 | F    | /lib/cryptsetup/checks/swap                                 |
| 2017-09-06 06:08:16 |    827 | F    | /lib/cryptsetup/checks/un_blkid                             |
| 2017-09-06 06:08:16 |    147 | F    | /lib/cryptsetup/checks/xfs                                  |
+---------------------+--------+------+-------------------------------------------------------------+

```
