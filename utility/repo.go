package utility

import (
	"github.com/platinasystems/go-common/logs"
	"github.com/platinasystems/pcc-blackbox/lib"
	"github.com/platinasystems/tiles/pccserver/models"
	"github.com/platinasystems/tiles/pccserver/utility"

	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type OsArchType uint8

const (
	OsUndefined OsArchType = iota
	OsDebian
	OsRedhat
	OsUnknown
)

func (a OsArchType) String() string {
	if a >= OsUnknown {
		return "unknown"
	}
	return []string{"undefined", "debian", "redhat"}[int(a)]
}

func OsArch() (os OsArchType, osVars map[string]string, err error) {
	osVars = map[string]string{}
	osRelease, e := Cat("/etc/os-release")
	if e != nil {
		err = fmt.Errorf("Cannot deterime os type: %v", err)
		return
	}
	return OsArchParse(string(osRelease))
}

func OsArchParse(osRelease string) (os OsArchType, osVars map[string]string, err error) {
	fields := []string{"ID", "VERSION", "VERSION_ID", "VERSION_CODENAME", "ID_LIKE", "PRETTY_NAME"}
	m := OsDataByFields(string(osRelease), fields)

	if m["ID"] == "debian" || m["ID"] == "ubuntu" || m["ID_LIKE"] == "debian" {
		os = OsDebian
	}

	if m["ID"] == "centos" || m["ID_LIKE"] == "rhel fedora" {
		os = OsRedhat
	}

	osVars = m

	return
}

func UpdatePkgs(isInstall bool, pkgs ...string) (err error) {
	maxRetries := 5
	retryInterval := 2 * time.Second
	maxWait := 10 * time.Second
	if len(pkgs) == 0 {
		return
	}
	osArch, _, e := OsArch()
	if e != nil {
		return e
	}
	args := []string{"-y"}
	if isInstall {
		args = append(args, "install")
	} else {
		args = append(args, "remove")
	}
	args = append(args, pkgs...)
	switch osArch {
	case OsDebian:
		for i := 0; i < maxRetries; i++ {
			if out, e := exec.Command("apt", args...).CombinedOutput(); e != nil {
				if !strings.Contains(string(out), "11: Resource temporarily unavailable") {
					err = fmt.Errorf(string(out))
					break
				} else {
					err = fmt.Errorf("After %v attempts %v", i+1, string(out))
				}
			} else {
				err = nil
				break
			}
			if i < maxRetries {
				time.Sleep(retryInterval)
			}
		}
	case OsRedhat:
		// yum does a retry automatically and hangs until pass/fail
		ctx, cancel := context.WithTimeout(context.Background(), maxWait)
		defer cancel()
		if out, e := exec.CommandContext(ctx, "yum", args...).CombinedOutput(); e != nil {
			err = fmt.Errorf("%v:%v", e, string(out))
		}
	}
	return
}

func UpdateRepo(isAdd bool, config models.RepoCombined) (err error) {
	osArch, _, e := OsArch()
	if e != nil {
		return e
	}
	switch osArch {
	case OsDebian:
		return UpdateRepoDeb(isAdd, config.Apt)
	case OsRedhat:
		return UpdateRepoRh(isAdd, config.Yum)
	default:
		err = fmt.Errorf("Cannot add repo to os type %v", osArch)
	}
	return
}

func UpdateRepoRh(isAdd bool, config models.RepoConfig) (err error) {
	if config.Source == "" {
		// nothing to do
		return
	}
	source := strings.TrimSpace(config.Source)
	if strings.HasSuffix(source, ".rpm") {
		return UpdateRPM(isAdd, source)
	}

	if config.FileName == "" {
		err = fmt.Errorf("no repo file name specified")
		return
	}

	if config.RepoId == "" {
		err = fmt.Errorf("no repo id specified")
		return
	}

	if config.RepoName == "" {
		config.RepoName = config.RepoId
	}

	fileName := strings.TrimSpace(config.FileName)
	if !strings.HasSuffix(fileName, ".repo") {
		fileName += ".repo"
	}
	repoName := strings.TrimSpace(config.RepoName)
	keyUrl := strings.TrimSpace(config.KeyUrl)

	repoPath := "/etc/yum.repos.d"
	filePath := filepath.Join(repoPath, fileName)

	clines := utility.TextLines{}
	id := fmt.Sprintf("[%v]", config.RepoId)
	clines.Append(id)
	clines.Append(fmt.Sprintf("name=%v", repoName))
	clines.Append(fmt.Sprintf("baseurl=%v", config.Source))
	keyPathBase := "/etc/pki/rpm-gpg/"
	if keyUrl != "" {
		clines.Append("gpgcheck=1")
		clines.Append(fmt.Sprintf("gpgkey=%v", keyUrl))

		// FIXME, download key and point to local file?
		if false {
			s := strings.Split(keyUrl, "/")
			keyName := s[len(s)-1]
			keyPath := filepath.Join(keyPathBase, keyName)
			if e := exec.Command("curl", keyUrl, "--output", keyPath).Run(); e == nil {
				clines.Append("gpgcheck=1")
				clines.Append(fmt.Sprintf("gpgkey=file://%v", keyPath))
			} else {
				err = fmt.Errorf("Failed to get gpgkey %v", keyUrl)
				return
			}
		}
	}
	clines.Append("enabled=1")

	// find any existing repo config
	oldFilePaths := map[string]bool{}
	GetFilesContaining(oldFilePaths, id, repoPath)

	// add if doesn't already exist
	if !oldFilePaths[filePath] && isAdd {
		b := []byte(clines.String() + "\n\n")
		if err = ioutil.WriteFile(filePath, b, 644); err == nil {
			oldFilePaths[filePath] = true
		} else {
			err = fmt.Errorf("Error writing %v: %v", filePath, err)
			return
		}
	}

	// mark all other instances, if any, for removal
	for fp := range oldFilePaths {
		if fp != filePath || !isAdd {
			oldFilePaths[fp] = false
		}
	}

	for fp, keepFile := range oldFilePaths {
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			continue
		}
		old := string(b)
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		newLines := utility.TextLines{}
		for i := 0; i < len(lines); i++ {
			line := strings.TrimSpace(lines[i])
			if isBracket(line) {
				keepLine := true
				if line == id {
					keepLine = false
				} else {
					newLines.Append(line)
				}
				// advance to next []
				i++
				for i < len(lines) {
					line := strings.TrimSpace(lines[i])
					if isBracket(line) {
						i--
						break
					}
					if keepLine {
						newLines.Append(line)
					}
					i++
				}
				if i >= len(line) {
					// end of file
					break
				}
			} else {
				newLines.Append(line)
			}
		}
		if keepFile {
			// add the new config at the end
			newLines.Append(clines...)
		}

		if len(newLines) == 0 {
			// remove the file if empty
			os.Remove(fp)
			continue
		}
		new := strings.TrimSpace(newLines.String()) + "\n"
		if old != new {
			b = []byte(new)
			err = ioutil.WriteFile(fp, b, 644)
		}
	}
	return
}

func isBracket(s string) bool {
	return strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")
}

func UpdateRepoDeb(isAdd bool, config models.RepoConfig) (err error) {
	if config.Source == "" {
		// nothing to do
		return
	}
	source := config.Source

	// set file path
	aptPath := "/etc/apt/"
	fileName := fileNameFromSource(source)
	if fileName == "" && isAdd {
		err = fmt.Errorf("could not find url")
		return
	}
	if config.FileName != "" {
		fileName = config.FileName
	}
	filePath := filepath.Join(aptPath, "sources.list.d", fileName)

	// if key is specified, add it
	if config.KeyUrl != "" && isAdd {
		// try url first
		tmpKeyFile := filepath.Join("/tmp/", fileName+".key")
		e := exec.Command("curl", config.KeyUrl, "--output", tmpKeyFile).Run()
		if e == nil {
			e = exec.Command("apt-key", "add", tmpKeyFile).Run()
		}
		if e != nil {
			// maybe key is already there, log and keep going
			log.AuctaLogger.Debugf("error adding key: %v", e)
		}
		os.Remove(tmpKeyFile)
	} else if config.KeyServer != "" && config.Key != "" {
		// no url, but has keyServer and keyId, use that
		if e := exec.Command("apt-key", "adv", "--keyserver", config.KeyServer,
			"--recv-keys", config.Key).Run(); e != nil {
			// maybe key is already there, log and keep going
			log.AuctaLogger.Debugf("error adding key: %v", e)
		}
	}

	// find any existing source
	oldFilePaths := map[string]bool{}
	GetFilesContaining(oldFilePaths, source, aptPath)

	// add if filePath doesn't already exist
	if !oldFilePaths[filePath] && isAdd {
		b := []byte(fmt.Sprintf("%v\n", source))
		if err = ioutil.WriteFile(filePath, b, 644); err == nil {
			oldFilePaths[filePath] = true
		} else {
			err = fmt.Errorf("Error writing %v: %v", filePath, err)
			return
		}
	}

	// mark all other instances of source, if any, for removal
	for fp := range oldFilePaths {
		if fp != filePath || !isAdd {
			oldFilePaths[fp] = false
		}
	}

	for fp, keep := range oldFilePaths {
		b, err := ioutil.ReadFile(fp)
		if err != nil {
			continue
		}
		old := string(b)
		lines := strings.Split(strings.TrimSpace(string(b)), "\n")
		newlines := []string{}
		kept := false
		for i := range lines {
			line := strings.TrimSpace(lines[i])
			// use Contains instead of equal in case there are comments
			if !strings.Contains(line, source) {
				newlines = append(newlines, line)
				continue
			}
			if keep && !kept {
				// just in case there are multiple instance of this
				// keep only one
				newlines = append(newlines, source)
				kept = true
			}
		}
		if len(newlines) == 0 {
			// remove the file if empty
			os.Remove(fp)
			continue
		}
		new := strings.Join(newlines, "\n") + "\n"
		if old != new {
			b = []byte(new)
			err = ioutil.WriteFile(fp, b, 644)
		}
	}
	return
}

func fileNameFromSource(source string) (name string) {
	var u string
	ss := strings.Split(source, " ")
	for _, s := range ss {
		if strings.HasPrefix(s, "http") ||
			(strings.Contains(s, ".") &&
				strings.Contains(s, "/")) {
			u = strings.TrimSpace(s)
			break
		}
	}
	if u == "" {
		return
	}
	u = strings.TrimPrefix(u, "http://")
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimSuffix(u, "/")
	u = strings.ReplaceAll(u, "/", ".")
	u = strings.ReplaceAll(u, ".", "_")
	name = u + ".list"
	return
}

func UpdateRPM(isAdd bool, rpm string) (err error) {
	if isAdd {
		if out, e := exec.Command("yum", "-y", "install", rpm).CombinedOutput(); e != nil {
			// ignore if error is "Nothing to do", basically just means already installed
			if !strings.Contains(string(out), "Error: Nothing to do") {
				err = fmt.Errorf(string(out))
			}
		}
		return
	}

	// strip out the url part if any
	s := strings.Split(rpm, "/")
	pkg := s[len(s)-1]
	// strip out .rpm if any
	pkg = strings.TrimSuffix(pkg, ".rpm")
	// ignore error, could be already removed
	if out, e := exec.Command("yum", "-y", "remove", pkg).CombinedOutput(); e != nil {
		err = fmt.Errorf(string(out))
	}
	return
}

func GetFilesContaining(m map[string]bool, text string, path string) {
	if m == nil {
		// m should not be nil
		return
	}
	if out, e := exec.Command("grep", "-rI", text, path).CombinedOutput(); e == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if !strings.Contains(line, text) ||
				!strings.Contains(line, path) {
				continue
			}
			s := strings.Split(line, ":")
			m[strings.TrimSpace(s[0])] = true
		}
	}
}

func OsDataByFields(osRelease string, fields []string) (result map[string]string) {
	result = map[string]string{}
	lines := strings.Split(osRelease, "\n")
	for i := range lines {
		line := utility.TextLine(lines[i])
		for f := range fields {
			field := strings.Trim(fields[f], "=")
			prefix := field + "="
			if line.HasPrefix(prefix) {
				s := strings.Split(string(line), prefix)
				if len(s) > 1 {
					result[field] = strings.Trim(s[1], "\"")
				}
			}
		}
	}
	return
}

func DpkgGetVersion(pkg string) (v string, err error) {
	var out []byte
	format := `-f=${Version}`
	out, err = exec.Command("dpkg-query", "-W", format, pkg).Output()
	v = strings.TrimSpace(string(out))
	return
}

func DpkgStatus(pkg string) (status string, installed bool, err error) {
	var out []byte
	format := `-f=${Status}`
	out, err = exec.Command("dpkg-query", "-W", format, pkg).Output()
	status = strings.TrimSpace(string(out))
	installed = status == "install ok installed"
	return
}

func UpdatePkgsRemote(isInstall bool, pkgs ...string) (err error) {
	maxRetries := 5
	retryInterval := 2 * time.Second
	maxWait := 10 * time.Second
	if len(pkgs) == 0 {
		return
	}
	osArch, _, e := OsArch()
	if e != nil {
		return e
	}
	args := []string{"-y"}
	if isInstall {
		args = append(args, "install")
	} else {
		args = append(args, "remove")
	}
	args = append(args, pkgs...)
	switch osArch {
	case OsDebian:
		for i := 0; i < maxRetries; i++ {
			if out, e := exec.Command("apt", args...).CombinedOutput(); e != nil {
				if !strings.Contains(string(out), "11: Resource temporarily unavailable") {
					err = fmt.Errorf(string(out))
					break
				} else {
					err = fmt.Errorf("After %v attempts %v", i+1, string(out))
				}
			} else {
				err = nil
				break
			}
			if i < maxRetries {
				time.Sleep(retryInterval)
			}
		}
	case OsRedhat:
		// yum does a retry automatically and hangs until pass/fail
		ctx, cancel := context.WithTimeout(context.Background(), maxWait)
		defer cancel()
		if out, e := exec.CommandContext(ctx, "yum", args...).CombinedOutput(); e != nil {
			err = fmt.Errorf("%v:%v", e, string(out))
		}
	}
	return
}

func OsArchRemote(host string, sshC *pcc.SshConfiguration) (os OsArchType, osVars map[string]string, err error) {
	cmd := pcc.Cmd{
		Input: "cat /etc/os-release",
	}
	if err = sshC.Run(host, &cmd); err != nil {
		return
	}
	return OsArchParse(cmd.Stdout)
}
