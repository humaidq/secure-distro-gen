package builder

import (
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

	_, err = exec.Command("chmod", "+w", manifest).Output()
	if err != nil {
		return errors.Wrap(err, "chmod filesystem manifest")
	}

	{ // write filesystem.manifest
		output, err := exec.Command("proot",
			"-R", sess.chrootDir+"/",
			"-w", "/",
			"-b", "/proc/",
			"-b", "/dev/",
			"-b", "/sys/",
			"-0",
			"dpkg-query -W --showformat='${Package} ${Version}\n'").Output()

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

	_, err = exec.Command("cp", manifest, manifest+"-desktop").Output()
	if err != nil {
		return errors.Wrap(err, "copy filesystem manifest (desktop)")
	}

	_, err = exec.Command("mksquashfs", sess.chrootDir, manifest).Output()
	if err != nil {
		return errors.Wrap(err, "mksquashfs")
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

				hashes.WriteString(string(hash))
				hashes.WriteRune('\n')
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

	xorriso := exec.Command("xorriso", "-as", "mkisofs",
		"-r", "-V", sess.cust.DistName+" "+sess.cust.DistVer+" amd64",
		"--protective-msdos-label",
		"-b", "isolinux/isolinux.bin",
		"-no-emul-boot", "-boot-load-size", "4", "-boot-info-table",
		"--grub2-boot-info",
		"--grub2-mbr", "/usr/lib/grub/i386-pc/boot_hybrid.img",
		"--efi-boot", "boot/grub/efi.img", "--efi-boot-part", "--efi-boot-image",
		"-o", sess.tempDir+"/output.iso",
	)

	xorriso.Dir = sess.extractDir

	out, err := xorriso.Output()
	if err != nil {
		return errors.Wrap(err, "xorriso")
	}

	config.Logger.Debug(out)

	return nil
}
