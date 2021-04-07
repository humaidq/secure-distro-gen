package builder

import (
	"bytes"
	"os"
	"os/exec"

	"git.sr.ht/~humaid/linux-gen/config"

	"github.com/pkg/errors"
)

func execc(wd string, cmd string, args ...string) (string, error) {
	var bOut, bErr bytes.Buffer
	c := exec.Command(cmd, args...)
	c.Stderr = &bErr
	c.Stdout = &bOut
	c.Stderr = os.Stdout
	c.Stdout = os.Stdout
	// fix bug with "terminated with signal 11"
	// https://github.com/proot-me/proot/issues/106
	c.Env = append(c.Env, "PROOT_NO_SECCOMP=1")
	if wd != "" {
		c.Dir = wd
	}
	err := c.Run()
	if err != nil {
		return bOut.String() + ";" + bErr.String(), err
	}
	return bOut.String() + ";" + bErr.String(), nil
}

// extract mounts the original ISO file and copies the contents out of it. It
// also unsquashes the filesystem in the ISO.
// At the end of the function, it unmounts the ISO file and removes its
// directory.
func extract(sess *buildSession) error {
	var err error

	mkdir(sess.mountDir)
	mkdir(sess.extractDir)

	config.Logger.Debug("extract: mount ISO")
	// mount the ISO file
	_, err = exec.Command("mount", "-o", "loop", config.Config.OrigISOFile,
		sess.mountDir).Output()
	if err != nil {
		return errors.Wrap(err, "mount the ISO")
	}

	config.Logger.Debug(sess.mountDir, sess.extractDir)
	config.Logger.Debug("extract: rsync")

	// copy contents to extract folder
	o, err := execc("", "rsync", "--exclude=/casper/filesystem.squashfs",
		"-a", sess.mountDir+"/", sess.extractDir)
	if err != nil {
		config.Logger.Debug("copy contents (rsync)", o)
		//return errors.Wrap(err, "copy contents (rsync) "+o)
	}

	config.Logger.Debug("extract: unsquashfs")
	o, err = execc(sess.tempDir, "unsquashfs",
		sess.mountDir+"/casper/filesystem.squashfs")
	if err != nil {
		return errors.Wrap(err, "unsquashfs: "+o)
	}

	os.Rename(sess.tempDir+"/squashfs-root", sess.chrootDir)

	// unmount the file
	err = umount(sess.mountDir)
	if err != nil {
		return errors.Wrap(err, "umount")
	}

	err = os.RemoveAll(sess.mountDir)
	if err != nil {
		return errors.Wrap(err, "remove mount directory")
	}

	return nil
}
