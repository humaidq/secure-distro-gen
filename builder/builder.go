package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"git.sr.ht/~humaid/linux-gen/config"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// MOUNT is the mount directory for the ISO file.
const MOUNT = "/mnt"

// EXTRACT is where the ISO will be extracted to.
const EXTRACT = "/extract-cd"

// CHROOT is chroot environment for editing the system.
const CHROOT = "/edit"

type buildSession struct {
	tempDir                         string
	mountDir, extractDir, chrootDir string
	cust                            Customisation
}

// Customisation contains all the customisable parts of a Linux distribution.
type Customisation struct {
	Author            string
	DistName, DistVer string
	AddPackages       []string
	RemovePackages    []string
	TZ, Kbd           string
}

type SystemPackage struct {
	Name, Version string
	Section       string
	Installed     bool
}

type SystemMetadata struct {
	Packages  []SystemPackage
	Timezones []string
}

func GetMetadata() (SystemMetadata, error) {
	var err error

	if err = DependencyCheck(); err != nil {
		return SystemMetadata{}, err
	}

	dir, _ := os.Getwd()
	dir = dir + "/temp"
	sess := buildSession{
		tempDir:    dir,
		mountDir:   dir + MOUNT,
		extractDir: dir + EXTRACT,
		chrootDir:  dir + CHROOT,
	}

	if _, err := os.Stat(dir); err == nil {
		cleanup(&sess)
	}
	mkdir(dir)

	config.Logger.Debug("start extract")
	err = extract(&sess)
	if err != nil {
		config.Logger.Error("failed to extract",
			zap.String("tempDir", dir),
			zap.Error(err),
		)
		cleanup(&sess)
		return SystemMetadata{}, err
	}

	installed := make(map[string]bool)

	output, err := execc(sess.tempDir, "chroot", sess.chrootDir, "apt",
		"list", "--installed")
	if err != nil {
		fmt.Println(err)
		return SystemMetadata{}, errors.Wrap(err, "apt list")
	}
	for i, l := range strings.Split(output, "\n") {
		if i == 0 {
			continue // skip the "Listing..." line
		}

		sp := strings.Split(l, " ")
		if len(sp) < 2 {
			continue
		}

		pkgName := strings.Split(l, "/")[0]
		installed[pkgName] = true
	}

	output, err = execc(sess.tempDir, "chroot", sess.chrootDir, "apt",
		"list")
	if err != nil {
		fmt.Println(err)
		return SystemMetadata{}, errors.Wrap(err, "apt list")
	}

	sys := SystemMetadata{}

	for i, l := range strings.Split(output, "\n") {
		if i == 0 {
			continue // skip the "Listing..." line
		}
		sp := strings.Split(l, " ")
		if len(sp) < 2 {
			fmt.Println("LINE SMOL", l)
			continue
		}

		pkgName := strings.Split(l, "/")[0]
		pkgVer := sp[1]

		// clean up ver
		pkgVer = strings.Split(pkgVer, "-")[0]
		pkgVer = strings.Split(pkgVer, "~")[0]
		pkgVer = strings.Split(pkgVer, "+")[0]

		//fmt.Println(pkgName, pkgVer)

		sys.Packages = append(sys.Packages, SystemPackage{
			Name:      pkgName,
			Version:   pkgVer,
			Installed: installed[pkgName],
		})
	}

	fmt.Printf("We have %d packages\n", len(sys.Packages))
	fmt.Println("Getting info for each package")
	for i, p := range sys.Packages {
		output, err = execc(sess.tempDir, "chroot", sess.chrootDir, "apt-cache",
			"show", p.Name)

		if err == nil {
			part := strings.Split(output, "Section: ")[1]
			part = strings.Split(part, "\n")[0]

			sys.Packages[i].Section = part
			//fmt.Println("part", part)
		} else {
			fmt.Println("Cannot get info for package", p.Name)
		}

		if i%50 == 0 {
			fmt.Printf("%d/%d complete\n", i, len(sys.Packages))
		}
	}

	cleanup(&sess)

	b, err := json.Marshal(sys)
	if err != nil {
		panic(err)
	}

	writeToFile("./metadata.json", string(b))
	fmt.Println("Written to metadata file")

	return sys, nil
}

// Start customises a Linux distribution based on the given customisation.
// Returns the location of the ISO, or an error.
func Start(cust Customisation) (string, error) {
	var err error

	if err = DependencyCheck(); err != nil {
		return "", err
	}

	/*dir, err := ioutil.TempDir("", "linux-gen")
	if err != nil {
		config.Logger.Error("failed to create temp dir",
			zap.String("author", cust.Author),
			zap.String("tempDir", dir),
			zap.Error(err),
		)
		return "", err
	}*/
	dir, _ := os.Getwd()
	dir = dir + "/temp"
	sess := buildSession{
		tempDir:    dir,
		mountDir:   dir + MOUNT,
		extractDir: dir + EXTRACT,
		chrootDir:  dir + CHROOT,
		cust:       cust,
	}

	if _, err := os.Stat(dir); err == nil {
		cleanup(&sess)
	}
	mkdir(dir)

	config.Logger.Debug("start extract")
	err = extract(&sess)
	if err != nil {
		config.Logger.Error("failed to extract",
			zap.String("author", cust.Author),
			zap.String("tempDir", dir),
			zap.Error(err),
		)
		cleanup(&sess)
		return "", err
	}

	config.Logger.Debug("start customise")
	err = customise(&sess)
	if err != nil {
		config.Logger.Error("failed to customise",
			zap.String("author", cust.Author),
			zap.String("tempDir", dir),
			zap.Error(err),
		)
		cleanup(&sess)
		return "", err
	}

	config.Logger.Debug("start build")
	err = build(&sess)
	if err != nil {
		config.Logger.Error("failed to build",
			zap.String("author", cust.Author),
			zap.String("tempDir", dir),
			zap.Error(err),
		)
		cleanup(&sess)
		return "", err
	}

	return "", nil
}

// cleanup cleans up the build session in case of unexpected failure.
// Only to be used when the build session unexpectedly fails!
func cleanup(sess *buildSession) {
	config.Logger.Debug("running cleanup")
	// First, we'll umount everything we can (so we don't cause any damage)
	if isMountpoint(sess.chrootDir + "/dev") { // runs after umounting everyhing in mountDir
		config.Logger.Debug("/dev is mounted, umounting")
		e := umount(sess.chrootDir + "/dev")
		if e != nil {
			panic("Cannot umount /dev!")
		}
	}

	if isMountpoint(sess.mountDir) { // runs after umounting everyhing in mountDir
		config.Logger.Debug("mountDir is mounted, umounting")
		e := umount(sess.mountDir)
		if e != nil {
			panic("Cannot umount /dev!")
		}
	}

	config.Logger.Debug("Will remove all when press enter")
	fmt.Scanln()
	os.RemoveAll(sess.tempDir)
}

// umount the given directory. May return an error.
func umount(dir string) (err error) {
	_, err = exec.Command("umount", dir).Output()
	return
}

// isMointpoint returns whether the directory given is a current mountpoint.
func isMountpoint(dir string) bool {
	_, err := exec.Command("mountpoint", dir).Output()
	return err == nil
}

// mkdir simply makes a directory with 0o777. May return an error.
func mkdir(dir string) error {
	return os.Mkdir(dir, os.ModePerm)
}

// writeToFile writes given text in the file provided, replacing existing
// contents. May return an error.
func writeToFile(file string, text string) error {
	f, err := os.Create(file)
	if err != nil {
		return errors.Wrap(err, "create file")
	}
	defer f.Close()

	_, err = io.WriteString(f, text)
	if err != nil {
		return errors.Wrap(err, "write string")
	}
	err = f.Sync()
	if err != nil {
		return errors.Wrap(err, "sync file")
	}

	return nil
}

func proot(sess *buildSession, comm string, args string) (output string, err error) {
	output, err = execc(sess.tempDir, "proot",
		"-R", sess.chrootDir+"/",
		"-w", "/",
		"-b", "/proc/",
		"-b", "/dev/",
		"-b", "/sys/",
		"-0", comm, args)
	fmt.Println("proot output:", output)
	return
}
