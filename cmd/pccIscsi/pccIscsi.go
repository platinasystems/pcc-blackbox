package main

import (
	"github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/pcc-blackbox/utility"

	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

const (
	rotational = "rotational"
	solidstate = "solidstate"
	bpath      = "/srv/iscsi"
	cpath      = "/etc/tgt/conf.d"
	ipath      = "/etc/iscsi/nodes"
	user       = "platina"
	passwd     = "pf00itn2"
)

type sshConfig struct {
	pcc.SshConfiguration
	host   string
	osArch utility.OsArchType
}

func InstallDependencies() (err error) {
	osArch, _, _ := utility.OsArch()
	switch osArch {
	case utility.OsDebian:
		if _, installed, _ := utility.DpkgStatus("tgt"); !installed {
			fmt.Println("Installing package tgt")
			utility.UpdatePkgs(true, "tgt")
		}
		if _, installed, _ := utility.DpkgStatus("open-iscsi"); !installed {
			fmt.Println("Installing package open-iscsi")
			utility.UpdatePkgs(true, "open-iscsi")
		}
	case utility.OsRedhat:
		utility.UpdatePkgs(true, "scsi-target-utils", "iscsi-initiator-utils")
	default:
		err = fmt.Errorf("Unknown OS type %v", osArch)
		return
	}
	return
}

func (sshC *sshConfig) Execute(cmds ...*pcc.Cmd) (err error) {
	return sshC.Run(sshC.host, cmds...)
}

func (sshC *sshConfig) OsArch() (os utility.OsArchType, osVars map[string]string, err error) {
	cmd := pcc.Cmd{
		Input: "cat /etc/os-release",
	}
	if err = sshC.Execute(&cmd); err != nil {
		return
	}
	os, osVars, err = utility.OsArchParse(cmd.Stdout)
	sshC.osArch = os
	return
}

func (sshC *sshConfig) MaybeOsArch() {
	if sshC.osArch == utility.OsUndefined {
		sshC.OsArch()
	}
}

func (sshC *sshConfig) UpdatePkgs(isInstall bool, pkgs ...string) (err error) {
	var c string
	sshC.MaybeOsArch()
	if len(pkgs) == 0 {
		return
	}
	args := []string{"-y"}
	if isInstall {
		args = append(args, "install")
	} else {
		args = append(args, "remove")
	}
	args = append(args, pkgs...)

	switch sshC.osArch {
	case utility.OsDebian:
		c = "sudo apt"
	case utility.OsRedhat:
		c = "sudo yum"
	default:
		err = fmt.Errorf("Unknown os type %v", sshC.osArch)
		return
	}

	for _, arg := range args {
		c += " " + arg
	}
	cmd := pcc.Cmd{
		Input: c,
	}

	err = sshC.Execute(&cmd)

	return
}

// only accept M or G suffix for now
func ValidateSize(n int, size string) (err error) {
	var (
		s string
		i int
		b syscall.Statfs_t
	)
	if strings.HasSuffix(size, "M") {
		s = strings.TrimSuffix(size, "M")
		i, _ = strconv.Atoi(s)
		i = i * n * 1000 * 1000
	}
	if strings.HasSuffix(size, "G") {
		s = strings.TrimSuffix(size, "G")
		i, _ = strconv.Atoi(s)
		i = i * n * 1000 * 1000 * 1000
	}
	if i <= 0 {
		err = fmt.Errorf("size expected with suffix M or G: e.g. 100M, 10G")
		return
	}

	syscall.Statfs("/", &b)
	free := b.Bfree * uint64(b.Bsize)
	if int(free)-i < 10*1000*1000*1000 {
		err = fmt.Errorf("Not enough disk space to create %v drives of size %v",
			n, size)
	}
	return
}

func RemoveTgt(n int, size, driveType string) (errs utility.Errors) {
	var bs []string
	// find matching drives
	bpath := "/srv/iscsi"
	fs, err := ioutil.ReadDir(bpath)
	if err != nil {
		errs.Append(err)
		return
	}
	prefix := fmt.Sprintf("platina-%v-%v-vol", driveType, size)
	suffix := fmt.Sprintf("-sparse.raw")
	for _, f := range fs {
		bname := f.Name()
		if strings.HasPrefix(bname, prefix) && strings.HasSuffix(bname, suffix) {
			bs = append(bs, bname)
		}
	}
	if len(bs) < n {
		// delete them all
		for _, bname := range bs {
			bfp := filepath.Join(bpath, bname)
			fmt.Println("Removing", bfp)
			os.Remove(bfp)
			cname := cnameFromBname(bname)
			cfp := filepath.Join(cpath, cname)
			fmt.Println("Removing", cfp)
			os.Remove(cfp)
		}
		return
	}
	// sort high to low and delete n from high down
	sort.Slice(bs, func(i, j int) bool {
		return volNum(bs[i]) > volNum(bs[j])
	})
	for i := 0; i < n; i++ {
		bfp := filepath.Join(bpath, bs[i])
		fmt.Println("Removing", bfp)
		os.Remove(bfp)
		cname := cnameFromBname(bs[i])
		cfp := filepath.Join(cpath, cname)
		fmt.Println("Removing", cfp)
		os.Remove(cfp)
	}
	return
}

func volNum(name string) (v int) {
	fs := strings.Split(name, "-")
	for _, s := range fs {
		if strings.HasPrefix(s, "vol") {
			s2 := strings.TrimPrefix(s, "vol")
			v, _ = strconv.Atoi(s2)
		}
	}
	return
}

func CreateTgt(n int, size, driveType string) (errs utility.Errors) {
	// create tgt conf
	os.Mkdir(bpath, 755)
	for i := 1; i <= n; i++ {
		// create backing store first if doesn't exist
		bname := fmt.Sprintf("platina-%v-%v-vol%v-sparse.raw", driveType, size, i)
		bfp := filepath.Join(bpath, bname)
		fmt.Println("Creating", bfp)
		if _, e := os.Stat(bfp); os.IsNotExist(e) {
			if out, e := exec.Command("dd", "if=/dev/zero", "bs=1", "count=1", "seek="+size, "of="+bfp).CombinedOutput(); e != nil {
				e = fmt.Errorf("%v: %v", e, string(out))
				errs.Append(e)
				continue

			}
		}

		cname := cnameFromBname(bname)
		lines := utility.TextLines{}
		lines.Append(fmt.Sprintf("<target %v>", iqn(driveType, size, i)))
		lines.Append(fmt.Sprintf("    backing-store /srv/iscsi/%v", bname))
		lines.Append(fmt.Sprintf("    initiator-address 127.0.0.1"))
		lines.Append(fmt.Sprintf("    incominguser %v %v", user, passwd))
		lines.Append(fmt.Sprintf("    vendor_id platina"))
		lines.Append(fmt.Sprintf("    product_id platina-%v", driveType))
		lines.Append(fmt.Sprintf("</target>"))
		cfp := filepath.Join(cpath, cname)
		fmt.Println("Creating", cfp)
		if err := ioutil.WriteFile(cfp, []byte(lines.String()), 0644); err != nil {
			err = fmt.Errorf("Error creating %v: %v", cfp, err)
			errs.Append(err)
		}
	}
	return
}

func RemoveInitiators(n int, size, driveType string) (errs utility.Errors) {
	for i := 1; i <= n; i++ {
		iname := iqn(driveType, size, i)
		errs.Append(exec.Command("iscsiadm", "-m", "node", "-T", iname, "-p", "127.0.0.1:3260", "-u").Run())
		errs.Append(exec.Command("iscsiadm", "-m", "node", "-o", "delete", "-T", iname).Run())
	}
	return
}

func CreateInitiators(n int, size, driveType string) (errs utility.Errors) {
	for i := 1; i <= n; i++ {
		iname := iqn(driveType, size, i)
		dirP := filepath.Join(ipath, iname)
		fs, err := ioutil.ReadDir(dirP)
		if err != nil {
			continue
		}
		// usually just 1 subdir
		for _, f := range fs {
			if !f.IsDir() {
				continue
			}
			fp := filepath.Join(dirP, f.Name(), "default")
			out, err := ioutil.ReadFile(fp)
			if err != nil {
				continue
			}
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			newlines := []string{}
			for _, line := range lines {
				if strings.HasPrefix(line, "node.session.auth.authmethod") {
					newlines = append(newlines, "node.session.auth.authmethod = CHAP")
					newlines = append(newlines, fmt.Sprintf("node.session.auth.username = %v", user))
					newlines = append(newlines, fmt.Sprintf("node.session.auth.password = %v", passwd))
					continue
				}
				if strings.HasPrefix(line, "node.startup") {
					newlines = append(newlines, "node.startup = automatic")
					continue
				}
				if strings.HasPrefix(line, "node.session.auth.username") {
					continue
				}
				if strings.HasPrefix(line, "node.session.auth.password") {
					continue
				}
				newlines = append(newlines, line)
			}
			new := strings.Join(newlines, "\n") + "\n"
			fmt.Println("Modifying", fp)
			ioutil.WriteFile(fp, []byte(new), 0644)
		}
	}
	return
}

func cnameFromBname(bname string) (cname string) {
	cname = strings.TrimSuffix(bname, "-sparse.raw")
	cname = cname + ".conf"
	return
}

func iqn(driveType string, size string, i int) string {
	return fmt.Sprintf("iqn.2021-04.%v.%v.platina.io:vol%v", driveType, size, i)
}

func main() {
	var (
		nr, ns                 int
		host, user, keyPath    string
		remote, create, remove bool
		sshC                   sshConfig
	)

	selfpath, err := os.Readlink("/proc/self/exe")
	self := filepath.Base(selfpath)
	if err != nil {
		fmt.Println(err)
		return
	}

	user = "pcc"
	keyPath = "~/.ssh/id_rsa"

	size := "1G"
	usage := fmt.Sprintf("%v [-h hostIp|hostName] [--key sshKeyPath] [--user sshUser] [--create|--remove] [--rotational num_drives] [--solidstate num_drives] [--size size]",
		os.Args[0])
	usage += fmt.Sprintf("\nIf executing local, no need for -h, --key, or --user")
	for i, arg := range os.Args {
		if arg == "--create" {
			create = true
			continue
		}
		if arg == "--remove" {
			remove = true
			continue
		}

		if i+1 >= len(os.Args) {
			continue
		}

		switch {
		case arg == "-h":
			host = os.Args[i+1]
			remote = true
		case arg == "-u" || arg == "--user":
			user = os.Args[i+1]
		case arg == "-k" || arg == "--key":
			keyPath = os.Args[i+1]
		case arg == "-r" || arg == "--rotational":
			if n, _ := strconv.Atoi(os.Args[i+1]); n <= 0 {
				fmt.Println(usage)
				fmt.Println("num_drives must be an integer greater than 0")
				return
			} else {
				nr = n
			}
		case (arg == "-ss" || arg == "--solidstate"):
			if n, _ := strconv.Atoi(os.Args[i+1]); n <= 0 {
				fmt.Println(usage)
				fmt.Println("num_drives must be an integer greater than 0")
				return
			} else {
				ns = n
			}
		case arg == "--size":
			s := os.Args[i+1]
			if !strings.HasSuffix(s, "M") && !strings.HasSuffix(s, "G") {
				fmt.Println(usage)
				fmt.Println("size expected with suffix M or G: e.g. 100M, 10G")
				return
			}
			size = s
		}
	}
	if (nr == 0 && ns == 0) ||
		(!create && !remove) {
		fmt.Println(usage)
		return
	}
	if err := ValidateSize(nr+ns, size); err != nil {
		fmt.Println(err)
	}

	prefix := "Creating"
	if remove {
		prefix = "Removing"
	}
	if remote {
		fmt.Printf("%v remotely on %v\n", prefix, host)
		sshC.host = host
		sshC.User = user
		sshC.Key = keyPath
		sshC.Port = 22
		if err := sshC.Init(); err != nil {
			fmt.Println(err)
			return
		}
		destpath := fmt.Sprintf("/tmp/%v", self)
		dest := fmt.Sprintf("%v@%v:%v", user, host, destpath)
		if out, e := exec.Command("scp", "-i", keyPath,
			"-o", "StrictHostKeyChecking=no",
			"-o", " UserKnownHostsFile=/dev/null",
			selfpath, dest).CombinedOutput(); e != nil {
			err = fmt.Errorf("error copying %v to %v: %v", selfpath, dest, string(out))
		}
		defer func() {
			cmd := pcc.Cmd{Input: fmt.Sprintf("sudo rm %v", destpath)}
			if err := sshC.Execute(&cmd); err != nil {
				fmt.Println("Error on remote node:", err)
			}
			if cmd.Stderr != "" {
				fmt.Println("Error on remote node:", cmd.Stderr)
			}
		}()

		rArgs := fmt.Sprintf("sudo %v", destpath)
		for i, arg := range os.Args {
			if i == 0 || arg == "-h" || arg == host ||
				arg == "--key" || arg == "-k" || arg == keyPath ||
				arg == "--user" || arg == "-u" || arg == user {
				continue
			}
			rArgs += " " + arg
		}
		cmd := pcc.Cmd{
			Input: rArgs,
		}
		fmt.Println(cmd.Input)
		if err := sshC.Execute(&cmd); err == nil && cmd.Stderr == "" {
			fmt.Println(cmd.Stdout)
		} else {
			fmt.Println("Error on remote node:", err)
		}
		if cmd.Stderr != "" {
			fmt.Println("Error on remote node:", cmd.Stderr)
		}
		return
	}

	// Following executes locally
	errs := utility.Errors{}
	if nr > 0 {
		if create {
			errs.Append(CreateTgt(nr, size, rotational)...)
		}
		if remove {
			errs.Append(RemoveTgt(nr, size, rotational)...)
		}
	}
	if ns > 0 {
		if create {
			errs.Append(CreateTgt(ns, size, solidstate)...)
		}
		if remove {
			errs.Append(RemoveTgt(ns, size, solidstate)...)
		}
	}
	if !errs.IsEmpty() {
		fmt.Println(errs)
		return
	}
	fmt.Println("systemctl restart tgt")
	if out, err := exec.Command("systemctl", "restart", "tgt").CombinedOutput(); err != nil {
		fmt.Println("%v: %v", err, string(out))
	}
	out, _ := exec.Command("tgtadm", "--mode", "target", "--op", "show").Output()
	for _, line := range utility.Egrep(out, "Target", "Backing store type", "Backing store path") {
		fmt.Println(line)
	}

	// initiators
	if create {
		// do this once only because it will replace the contents of /etc/iscsi/nodes with default
		out, err := exec.Command("iscsiadm", "-m", "discovery", "-t", "st", "-p", "127.0.0.1").CombinedOutput()
		if err != nil {
			fmt.Println("%v: %v", err, string(out))
			return
		}
	}

	if nr > 0 {
		if create {
			errs.Append(CreateInitiators(nr, size, rotational)...)
		}
		if remove {
			// don't err incase removing already removed or non-existent drive
			RemoveInitiators(nr, size, rotational)
		}
	}
	if ns > 0 {
		if create {
			errs.Append(CreateInitiators(ns, size, solidstate)...)
		}
		if remove {
			// don't err incase removing already removed or non-existent drive
			RemoveInitiators(ns, size, solidstate)
		}
	}
	if !errs.IsEmpty() {
		fmt.Println(errs)
		return
	}
	fmt.Println("systemctl restart open-iscsi")
	if out, err := exec.Command("systemctl", "restart", "open-iscsi").CombinedOutput(); err != nil {
		fmt.Println("%v: %v", err, string(out))
	}
	out2, _ := exec.Command("iscsiadm", "-m", "session", "-P", "3").CombinedOutput()
	for _, line := range utility.Egrep(out2, "Target:", "Attached scsi disk") {
		fmt.Println(line)
	}
}
