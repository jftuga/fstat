# fstat

`fstat` is useful when you need to obtain file names, sizes, and timestamps across multiple directories.  You can also sort the output by timestamp, file size, and file name (both case-sensitive and case-insensitive). 

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
       (filename should contain a list of files)

  -M	add milliseconds to file time stamps
  -c	add comma thousands separator to file sizes
  -f string
    	use these files instead of from a file or STDIN, can include wildcards
  -id
    	include only directories
  -if
    	include only files
  -il
    	include only symbolic links
  -m	convert file sizes to mebibytes
  -oc
    	ouput to CSV format
  -oh
    	ouput to HTML format
  -oj
    	ouput to JSON format
  -q	do not display file errors
  -sD
    	sort by file modified date, newest first
  -sI
    	sort by file name, ignore case, reverse alphabetical order
  -sN
    	sort by file name, reverse alphabetical order
  -sS
    	sort by file size, descending
  -sd
    	sort by file modified date
  -si
    	sort by file name, ignore case
  -sn
    	sort by file name
  -ss
    	sort by file size
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
| 2018-12-05 10:10:48 |  289584 | F    | c:\Program Files\Microsoft Office\root\Office16\1033\VISEVMON.EXE                                 |
| 2019-02-20 14:34:56 |   40264 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\Common.DBConnection.exe                       |
| 2019-02-20 14:34:56 |  186704 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\DATABASECOMPARE.EXE                           |
| 2018-12-05 10:12:18 |  267384 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\filecompare.exe                               |
| 2019-02-20 14:34:56 |  465528 | F    | c:\Program Files\Microsoft Office\root\Office16\DCF\SPREADSHEETCOMPARE.EXE                        |
| 2018-12-05 10:11:37 |   82240 | F    | c:\Program Files\Microsoft Office\root\Office16\SkypeSrv\SKYPESERVER.EXE                          |
| 2019-01-10 10:06:19 |  372864 | F    | c:\Program Files\Microsoft Office\root\vfs\ProgramFilesX64\Microsoft Office\Office16\MSOHTMED.EXE |
+---------------------+---------+------+---------------------------------------------------------------------------------------------------+
```

Running `fstat` in Windows with `-f` option:
```
c:\> fstat.exe -f "c:\Windows\Microsoft.NET\Framework*\*\csc.exe"

+---------------------+---------+------+---------------------------------------------------------+
|      MOD TIME       |  SIZE   | TYPE |                          NAME                           |
+---------------------+---------+------+---------------------------------------------------------+
| 2016-05-25 10:56:04 | 1545864 | F    | c:\Windows\Microsoft.NET\Framework\v3.5\csc.exe         |
| 2017-04-21 17:53:36 | 2170488 | F    | c:\Windows\Microsoft.NET\Framework\v4.0.30319\csc.exe   |
| 2016-07-14 14:18:12 |   88712 | F    | c:\Windows\Microsoft.NET\Framework64\v2.0.50727\csc.exe |
| 2016-05-25 14:29:34 | 2288264 | F    | c:\Windows\Microsoft.NET\Framework64\v3.5\csc.exe       |
| 2017-04-21 17:50:55 | 2738296 | F    | c:\Windows\Microsoft.NET\Framework64\v4.0.30319\csc.exe |
| 2016-07-13 14:33:18 |   77960 | F    | c:\Windows\Microsoft.NET\Framework\v2.0.50727\csc.exe   |
+---------------------+---------+------+---------------------------------------------------------+
```

Running `fstat` on Linux, using `-f` option:
```
user@debian:~$ fstat -f "/usr/*bin/f*g /etc/pa*"

+---------------------+-------+------+--------------------+
|      MOD TIME       | SIZE  | TYPE |        NAME        |
+---------------------+-------+------+--------------------+
| 2018-12-23 18:59:35 |  1421 | F    | /etc/passwd        |
| 2018-12-23 18:59:35 |  1421 | F    | /etc/passwd-       |
| 2017-05-17 07:59:59 | 18728 | F    | /usr/bin/faillog   |
| 2017-01-31 19:54:55 | 14352 | F    | /usr/sbin/filefrag |
| 2017-05-27 11:44:02 |   552 | F    | /etc/pam.conf      |
| 2019-03-04 06:17:55 |  4096 | D    | /etc/pam.d         |
+---------------------+-------+------+--------------------+
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
| 2017-09-06 06:08:16 |   1414 | F    | /lib/cryptsetup/scripts/decrypt_opensc                      |
| 2017-09-06 06:08:16 |    347 | F    | /lib/cryptsetup/scripts/decrypt_ssl                         |
| 2017-09-06 06:08:16 |    387 | F    | /lib/cryptsetup/checks/ext2                                 |
| 2017-09-06 06:08:16 |    147 | F    | /lib/cryptsetup/checks/xfs                                  |
+---------------------+--------+------+-------------------------------------------------------------+

```
