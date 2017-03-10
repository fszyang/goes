// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mount

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const Name = "mount"

// hack around syscall incorrect definition
const MS_NOUSER uintptr = (1 << 31)
const procFilesystems = "/proc/filesystems"

type fstabEntry struct {
	fsSpec  string
	fsFile  string
	fsType  string
	mntOpts string
}

type fsType struct {
	name	string
	nodev	bool
}


type filesystems struct {
	isNoDev		map[string]bool
	autoList	[]string
}

type cmd struct{}

func New() cmd { return cmd{} }

var translations = []struct {
	name string
	bits uintptr
	set  bool
}{
	{"-read-only", syscall.MS_RDONLY, true},
	{"-read-write", syscall.MS_RDONLY, false},
	{"-suid", syscall.MS_NOSUID, false},
	{"-no-suid", syscall.MS_NOSUID, true},
	{"-dev", syscall.MS_NODEV, false},
	{"-no-dev", syscall.MS_NODEV, true},
	{"-exec", syscall.MS_NOEXEC, false},
	{"-no-exec", syscall.MS_NOEXEC, true},
	{"-synchronous", syscall.MS_SYNCHRONOUS, true},
	{"-no-synchronous", syscall.MS_SYNCHRONOUS, true},
	{"-remount", syscall.MS_REMOUNT, true},
	{"-mand", syscall.MS_MANDLOCK, true},
	{"-no-mand", syscall.MS_MANDLOCK, false},
	{"-dirsync", syscall.MS_DIRSYNC, true},
	{"-no-dirsync", syscall.MS_DIRSYNC, false},
	{"-atime", syscall.MS_NOATIME, false},
	{"-no-atime", syscall.MS_NOATIME, true},
	{"-diratime", syscall.MS_NODIRATIME, false},
	{"-no-diratime", syscall.MS_NODIRATIME, true},
	{"-bind", syscall.MS_BIND, true},
	{"-move", syscall.MS_MOVE, true},
	{"-silent", syscall.MS_SILENT, true},
	{"-loud", syscall.MS_SILENT, false},
	{"-posixacl", syscall.MS_POSIXACL, true},
	{"-no-posixacl", syscall.MS_POSIXACL, false},
	{"-bindable", syscall.MS_UNBINDABLE, false},
	{"-unbindable", syscall.MS_UNBINDABLE, true},
	{"-private", syscall.MS_PRIVATE, true},
	{"-slave", syscall.MS_SLAVE, true},
	{"-shared", syscall.MS_SHARED, true},
	{"-relatime", syscall.MS_RELATIME, true},
	{"-no-relatime", syscall.MS_RELATIME, false},
	{"-iversion", syscall.MS_I_VERSION, true},
	{"-no-iversion", syscall.MS_I_VERSION, false},
	{"-strictatime", syscall.MS_STRICTATIME, true},
	{"-no-strictatime", syscall.MS_STRICTATIME, false},
}

type MountResult struct {
	err	error
	dev	string
	fstype	string
	dir	string
	flag	flags.Flag
}
	
func (r *MountResult) String() string {
	if r.err != nil {
		return fmt.Sprintf("%s: %v", r.dev, r.err)
	}
	if r.flag["--fake"] {
		return fmt.Sprintf("Would mount %s type %s at %s", r.dev, r.fstype, r.dir)
	}
	if r.flag["-v"] {
		return fmt.Sprintf("Mounted %s type %s at %s", r.dev, r.fstype, r.dir)
	}
	return ""
}
	
func (r *MountResult) ShowResult() {
	s := r.String()
	if s != "" {
		fmt.Println(s)
	}
}

type superBlock interface {
}

type unknownSB struct {
}

func readSuperBlock(dev string) (superBlock, error) {
	return &unknownSB{}, nil
}

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + " [OPTION]... DEVICE [DIRECTORY]" }

func (cmd) Main(args ...string) error {
	flag, args := flags.New(args,
		"--fake",
		"-v",
		"-a",
		"-F",
		"-defaults",
		"-p",
		"-r",
		"-read-write",
		"-suid",
		"-no-suid",
		"-dev",
		"-no-dev",
		"-exec",
		"-no-exec",
		"-synchronous",
		"-no-synchronous",
		"-remount",
		"-mand",
		"-no-mand",
		"-dirsync",
		"-no-dirsync",
		"-atime",
		"-no-atime",
		"-diratime",
		"-no-diratime",
		"-bind",
		"-move",
		"-silent",
		"-loud",
		"-posixacl",
		"-no-posixacl",
		"-bindable",
		"-unbindable",
		"-private",
		"-slave",
		"-shared",
		"-relatime",
		"-no-relatime",
		"-iversion",
		"-no-iversion",
		"-strictatime",
		"-no-strictatime")
	parm, args := parms.New(args, "-match", "-o", "-t")
	if len(parm["-t"]) == 0 {
		parm["-t"] = "auto"
	}

	fs, err := getFilesystems()

	if flag["-a"] {
		err = fs.mountall(flag, parm)
	} else {
		switch len(args) {
		case 0:
			err = show()
		case 1:
			if flag["-p"] {
				err = fs.mountprobe(args[0], flag, parm)
			} else {
				err = fs.fstab(args[0], flag, parm)
			}
		case 2:
			r := fs.mountone(parm["-t"], args[0], args[1], flag,
				parm)
			r.ShowResult()
			err = r.err
		default:
			err = fmt.Errorf("%v: unexpected", args[2:])
		}
	}
	return err
}

func pollMountResults (c chan *MountResult) (i int) {
	for {
		select {
		case r := <-c:
			r.ShowResult()
			i++
		default:
			return i
		}
	}		
	return i
}

func flushMountResults (c chan *MountResult, complete, count int) () {
	for i := complete; i < count; i++ {
		r := <- c
		r.ShowResult()
	}
}

func (fs *filesystems) mountall(flag flags.Flag, parm parms.Parm) error {
	fstab, err := loadFstab()
	if err != nil {
		return err
	}

	count := len(fstab)
	cap := 1
	if flag["-F"] {
		cap = count
	}

	complete := 0
	rchan := make(chan *MountResult, cap)

	for _, x := range fstab {
		go fs.goMountone(x.fsType, x.fsSpec, x.fsFile, flag, parm, rchan)
		complete += pollMountResults(rchan)
	}

	flushMountResults(rchan, complete, count)
	return nil
}

func (fs *filesystems)mountprobe(mountpoint string, flag flags.Flag, parm parms.Parm) error {
	files, err := ioutil.ReadDir("/sys/block")
	if err != nil {
		return err
	}
	complete := 0
	cap := 1
	if flag["-F"] {
		cap = len(files)
	}
	rchan := make(chan *MountResult, cap)
	
	for _, file := range files {
		mp := mountpoint + "/" + file.Name()
		if _, err := os.Stat(mp); os.IsNotExist(err) {
			err := os.Mkdir(mp, os.FileMode(0555))
			if err != nil {
				fmt.Println("mkdir", mp, "err:", err)
				return err
			}
		}
		go fs.goMountone(parm["-t"], "/dev/" + file.Name(), mp, flag, parm, rchan)
		complete += pollMountResults(rchan)
	}

	flushMountResults(rchan, complete, len(files))
	return nil
}

func (fs *filesystems)fstab(name string, flag flags.Flag, parm parms.Parm) error {
	fstab, err := loadFstab()
	if err != nil {
		return err
	}
	for _, x := range fstab {
		if name == x.fsSpec || name == x.fsFile {
			r := fs.mountone(x.fsType, x.fsSpec, x.fsFile,
				flag, parm)
			r.ShowResult()
			return r.err
		}
	}
	return nil
}

func loadFstab() ([]fstabEntry, error) {
	f, err := os.Open("/etc/fstab")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var fstab []fstabEntry
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Index(line, "#") < 0 {
			fields := strings.Fields(line)
			fstab = append(fstab, fstabEntry{
				fsSpec:  fields[0],
				fsFile:  fields[1],
				fsType:  fields[2],
				mntOpts: fields[3],
			})
		}
	}
	return fstab, scanner.Err()
}

func (fs *filesystems) mountone(t, dev, dir string, flag flags.Flag, parm parms.Parm) *MountResult {
	var flags uintptr
	if flag["-defaults"] {
		//  rw, suid, dev, exec, auto, nouser, async
		flags &^= syscall.MS_RDONLY
		flags &^= syscall.MS_NOSUID
		flags &^= syscall.MS_NODEV
		flags &^= syscall.MS_NOEXEC
		if t == "" {
			t = "auto"
		}
		flags |= MS_NOUSER
		flags |= syscall.MS_ASYNC
	}
	for _, x := range translations {
		if flag[x.name] {
			if x.set {
				flags |= x.bits
			} else {
				flags &^= x.bits
			}
		}
	}
	if flag["--fake"] {
		return &MountResult{nil, dev, t, dir, flag}
	}

	tryTypes := []string{t}
	nodev := false
	if t == "auto" {
		tryTypes = fs.autoList
	} else {
		nodev = fs.isNoDev[t]
	}

	if !nodev {
		_, err := readSuperBlock(dev)
		if err != nil {
			return &MountResult{err, dev, t, dir, flag}
		}
	}
	
	var err error
	for _, t := range tryTypes {
		err = syscall.Mount(dev, dir, t, flags, parm["-o"])
		if err == nil {
			return &MountResult{err, dev, t, dir, flag}
		}
	}

	return &MountResult{err, dev, t, dir, flag}
}

func (fs *filesystems)goMountone(t, dev, dir string, flag flags.Flag, parm parms.Parm, c chan *MountResult) {
	c <- fs.mountone(t, dev, dir, flag, parm)
}
	
func show() error {
	f, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		fmt.Print(fields[0], " on ", fields[1], " type ", fields[2],
			"(", fields[3], ")\n")

	}
	return scanner.Err()
}

func getFilesystems() (fsPtr *filesystems, err error) {
	f, err := os.Open(procFilesystems)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fs filesystems
	fs.isNoDev = make(map[string]bool)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		nodev := false
		if strings.HasPrefix(line, "nodev") {
			nodev = true
			line = strings.TrimPrefix(line, "nodev")
		}
		line = strings.TrimSpace(line)
		fs.isNoDev[line] = nodev
		if !nodev {
			fs.autoList = append(fs.autoList, line)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "scan:", procFilesystems, err)
		return nil, err
	}
	return &fs, nil
}

func (cmd) Apropos() map[string]string {
	return map[string]string{
		"en_US.UTF-8": "activated a filesystem",
	}
}

func (cmd) Man() map[string]string {
	return map[string]string{
		"en_US.UTF-8": `NAME
	mount - activate a filesystem

SYNOPSIS
	mount [OPTION]... [DEVICE DIR]

DESCRIPTION
	Mount a filesystem on a target directory.

OPTIONS
	--fake
	-v		verbose
	-a		all [-match MATCH[,...]]
	-t FSTYPE[,...]
	-o FSOPT[,...]
	-F		run mounts in parallel
	-p MNTPOINT	Probe for devices and mount under MNTPOINT
			Creating directories, and naming mount points
			after the Linux device name.

	Where MATCH, FSTYPE and FSOPT are comma separated lists.

FSTYPE
	May be anything listed in /proc/filesystems; for example:
	sysfs, ramfs, proc, tmpfs, devtmpfs, debugfs, securityfs,
	sockfs, pipefs, devpts, hugetlbfs, pstore, mqueue, btrfs,
	ext2, ext3, ext4, nfs, nfs4, nfsd, aufs

FILESYSTEM INDEPENDENT FLAGS
	-defaults	-read-write -dev -exec -suid
	-r		read only
	-read-write
	-suid		Obey suid and sgid bits
	-no-suid	Ignore suid and sgid bits
	-dev		Allow use of special device files
	-no-dev		Disallow use of special device files
	-exec		Allow program execution
	-no-exec	Disallow program execution
	-synchronous	Writes are synced at once
	-no-synchronous	Writes aren't synced at once 
	-remount	Alter flags of mounted filesystem
	-mand		Allow mandatory locks
	-no-mand	Disallow mandatory locks
	-dirsync	Directory modifications are synchronous
	-no-dirsync	Directory modifications are asynchronous
	-atime		Update inode access times
	-no-atime	Don't update inode access-times
	-diratime	Update directory access-times
	-no-diratime	Don't update directory access times
	-bind		Bind a file or directory
	-move		Relocate an existing mount point
	-silent
	-loud
	-posixacl	Filesystem doesn't apply umask
	-no-posixacl	Filesystem applies umask
	-bindable	Make mount point able to be bind mounted
	-unbindable	Make mount point unable to be bind mounted
	-private	Change to private subtree
	-slave		Change to slave subtree
	-shared		Change to shared subtree
	-relatime	Update atime relative to mtime/ctime
	-no-relatime	Disable relatime
	-iversion	Update inode I-Version field
	-no-iversion	Don't update inode I-Version field
	-strictatime	Always perform atime updates
	-no-strictatime	May skip atime updates`,
	}
}
