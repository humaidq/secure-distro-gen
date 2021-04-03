package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"

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

// Start customises a Linux distribution based on the given customisation.
// Returns the location of the ISO, or an error.
func Start(cust Customisation) (string, error) {
	var err error

	if err = DependencyCheck(); err != nil {
		return "", err
	}

	dir, _ := os.Getwd()
	dir = dir + "/temp"
	if _, err := os.Stat(dir); err == nil {
		os.RemoveAll(dir)
	}
	mkdir(dir)
	/*dir, err := ioutil.TempDir("", "linux-gen")
	if err != nil {
		config.Logger.Error("failed to create temp dir",
			zap.String("author", cust.Author),
			zap.String("tempDir", dir),
			zap.Error(err),
		)
		return "", err
	}*/
	sess := buildSession{
		tempDir:    dir,
		mountDir:   dir + MOUNT,
		extractDir: dir + EXTRACT,
		chrootDir:  dir + CHROOT,
		cust:       cust,
	}

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

	if isMountpoint(sess.mountDir) { // runs after umounting everyhing in mountDir
		config.Logger.Debug("mountDir is mounted, umounting")
		umount(sess.mountDir)
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

func proot(sess *buildSession, comm string) (output string, err error) {
	output, err = execc(sess.tempDir, "proot",
		"-R", sess.chrootDir+"/",
		"-w", "/",
		"-b", "/proc/",
		"-b", "/dev/",
		"-b", "/sys/",
		"-0",
		comm)

	return
}
