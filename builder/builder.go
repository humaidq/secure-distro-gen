package builder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"git.sr.ht/~humaid/linux-gen/config"
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
	DistName, DistVer string
}

func Start(cust Customisation) {
	dir := os.TempDir()
	sess := buildSession{
		tempDir:    dir,
		mountDir:   dir + MOUNT,
		extractDir: dir + EXTRACT,
		chrootDir:  dir + CHROOT,
		cust:       cust,
	}

	extract(&sess)
}

func mkdir(dir string) {
	os.Mkdir(dir, os.ModePerm)
}

func extract(sess *buildSession) {
	var err error

	mkdir(sess.mountDir)

	// mount the ISO file
	_, err = exec.Command("mount", "-o", "loop", config.Config.OrigISOFile,
		sess.mountDir).Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	// copy contents to extract folder
	_, err = exec.Command("rsync", "--exclude=/casper/filesystem.squashfs",
		"-a", sess.mountDir+"/", sess.extractDir).Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = exec.Command("unsquashfs",
		sess.mountDir+"/casper/filesystem.squashfs").Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	os.Rename("squashfs-root", sess.chrootDir)

	// unmount the file
	_, err = exec.Command("umount", sess.mountDir).Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	os.RemoveAll(sess.mountDir)
}

func build(sess *buildSession) {
	var err error

	manifest := sess.extractDir + "/casper/filesystem.manifest"

	_, err = exec.Command("chmod", "+w", manifest).Output()
	if err != nil {
		fmt.Println(err)
		return
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
			return
		}
		defer f.Close()

		_, err = io.WriteString(f, string(output))
		if err != nil {
			fmt.Println(err)
			return
		}
		f.Sync()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	_, err = exec.Command("cp", manifest, manifest+"-desktop").Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, err = exec.Command("mksquashfs", sess.chrootDir, manifest).Output()
	if err != nil {
		fmt.Println(err)
		return
	}

	{ // write filesystem.size
		size, err := exec.Command("du", "-sx", "--block-size=1", sess.chrootDir).Output()
		if err != nil {
			fmt.Println(err)
			return
		}

		sizeStripped := strings.Split(string(size), " ")[0]
		f, err := os.Create(sess.extractDir + "/casper/filesystem.size")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		_, err = io.WriteString(f, sizeStripped)
		if err != nil {
			fmt.Println(err)
			return
		}
		f.Sync()
		if err != nil {
			fmt.Println(err)
			return
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
			fmt.Println(err)
			return
		}
		defer f.Close()

		_, err = io.WriteString(f, hashes.String())
		if err != nil {
			fmt.Println(err)
			return
		}
		f.Sync()
		if err != nil {
			fmt.Println(err)
			return
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
		fmt.Println(err)
		return
	}
	fmt.Println(out)
}
