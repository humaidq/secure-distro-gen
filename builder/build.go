package builder

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"git.sr.ht/~humaid/linux-gen/config"
)

// build creates the new ISO file. It also performs some routine tasks, such as
// creating the filesystem manifest and calculate file hashes.
func build(sess *buildSession) error {
	var err error

	manifest := sess.extractDir + "/casper/filesystem.manifest"
	config.Logger.Debug("build: chmod")
	_, err = exec.Command("chmod", "+w", manifest).Output()
	if err != nil {
		return errors.Wrap(err, "chmod filesystem manifest")
	}

	{ // write filesystem.manifest
		output, err := execc(sess.tempDir, "chroot", sess.chrootDir, "dpkg-query",
			"-W", "--showformat='${Package} ${Version}\\n'")
		if err != nil {
			fmt.Println(err)
			return errors.Wrap(err, "create dpkg query")
		}

		f, err := os.Create(sess.extractDir + "/casper/filesystem.manifest")
		if err != nil {
			fmt.Println(err)
			return errors.Wrap(err, "create filesystem manifest")
		}
		defer f.Close()

		_, err = io.WriteString(f, string(output))
		if err != nil {
			return errors.Wrap(err, "write filesystem manifest")
		}
		err = f.Sync()
		if err != nil {
			return errors.Wrap(err, "sync filesystem manifest")
		}
	}

	config.Logger.Debug("build: cp manifest")
	_, err = exec.Command("cp", manifest, manifest+"-desktop").Output()
	if err != nil {
		return errors.Wrap(err, "copy filesystem manifest (desktop)")
	}
	// TODO remove from the manifest (desktop) the calamares and casper
	// packages

	config.Logger.Debug("build: mksquashfs")
	o, err := execc("", "mksquashfs",
		sess.chrootDir, sess.extractDir+"/casper/filesystem.squashfs",
		"-comp", "lz4")
	if err != nil {
		return errors.Wrap(err, "mksquashfs "+string(o))
	}

	{ // write filesystem.size
		size, err := exec.Command("du", "-sx", "--block-size=1", sess.chrootDir).Output()
		if err != nil {
			return errors.Wrap(err, "du filesystem size")
		}

		sizeStripped := strings.Split(string(size), " ")[0]
		f, err := os.Create(sess.extractDir + "/casper/filesystem.size")
		if err != nil {
			return errors.Wrap(err, "create filesystem size")
		}
		defer f.Close()

		_, err = io.WriteString(f, sizeStripped)
		if err != nil {
			return errors.Wrap(err, "create filesystem size")
		}
		err = f.Sync()
		if err != nil {
			return errors.Wrap(err, "sync filesystem size")
		}
	}

	// Update md5sum.txt file
	os.Remove(sess.extractDir + "/md5sum.txt")

	config.Logger.Debug("build: hashes")
	var hashes strings.Builder

	err = filepath.Walk(sess.extractDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && path != sess.extractDir+"/isolinux/boot.cat" {
				hash, err := exec.Command("md5sum", path).Output()
				if err != nil {
					fmt.Println(err)
					return err
				}

				hashes.WriteString(strings.Replace(string(hash), sess.extractDir, "", -1))
			}
			return nil
		})

	{ // write file
		f, err := os.Create(sess.extractDir + "/md5sum.txt")
		if err != nil {
			return errors.Wrap(err, "create md5sum")
		}
		defer f.Close()

		_, err = io.WriteString(f, hashes.String())
		if err != nil {
			return errors.Wrap(err, "write md5sum")
		}
		err = f.Sync()
		if err != nil {
			return errors.Wrap(err, "sync md5sum")
		}
	}

	config.Logger.Debug("build: xorriso")
	xorriso := exec.Command("xorriso", "-as", "mkisofs",
		"-r", "-V", sess.cust.DistName+" "+sess.cust.DistVer+" amd64",
		"--protective-msdos-label",
		"-b", "isolinux/isolinux.bin",
		"-no-emul-boot", "-boot-load-size", "4", "-boot-info-table",
		"--grub2-boot-info",
		"--grub2-mbr", "/usr/lib/grub/i386-pc/boot_hybrid.img",
		"--efi-boot", "boot/grub/efi.img", "--efi-boot-part", "--efi-boot-image",
		"-o", sess.tempDir+"/output.iso", ".",
	)

	var bOut, bErr bytes.Buffer
	xorriso.Dir = sess.extractDir
	xorriso.Stderr = &bErr
	xorriso.Stdout = &bOut
	err = xorriso.Run()
	config.Logger.Debug("xor error: " + bErr.String())
	config.Logger.Debug("xor out: " + bOut.String())
	if err != nil {
		return errors.Wrap(err, "xorriso")
	}

	return nil
}
