package builder

import (
	"os"
	"os/exec"

	"git.sr.ht/~humaid/linux-gen/config"

	"github.com/pkg/errors"
)

// extract mounts the original ISO file and copies the contents out of it. It
// also unsquashes the filesystem in the ISO.
// At the end of the function, it unmounts the ISO file and removes its
// directory.
func extract(sess *buildSession) error {
	var err error

	mkdir(sess.mountDir)

	// mount the ISO file
	_, err = exec.Command("mount", "-o", "loop", config.Config.OrigISOFile,
		sess.mountDir).Output()
	if err != nil {
		return errors.Wrap(err, "mount the ISO")
	}

	// copy contents to extract folder
	_, err = exec.Command("rsync", "--exclude=/casper/filesystem.squashfs",
		"-a", sess.mountDir+"/", sess.extractDir).Output()
	if err != nil {
		return errors.Wrap(err, "copy contents (rsync)")
	}

	_, err = exec.Command("unsquashfs",
		sess.mountDir+"/casper/filesystem.squashfs").Output()
	if err != nil {
		return errors.Wrap(err, "unsquashfs")
	}

	os.Rename("squashfs-root", sess.chrootDir)

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
