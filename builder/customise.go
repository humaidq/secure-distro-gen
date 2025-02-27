package builder

import (
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// DATE is the ISO date layout without dashes.
const DATE = "20060102"

func sedFile(file string, pattern string) error {
	_, err := exec.Command("sed", "-i", "-e", pattern, file).Output()
	return err
}

// customise modifies the distribution filesystem and configurations to match
// the configuration provided in the build session.
func customise(sess *buildSession) error {
	var date = time.Now().Format(DATE)

	setProgress(sess, "Customising: Applying Branding", 6)

	sedFile(sess.extractDir+"/isolinux/txt.cfg", "s/Lubuntu/"+sess.cust.DistName+"/g")
	sedFile(sess.extractDir+"/boot/grub/grub.cfg", "s/Lubuntu/"+sess.cust.DistName+"/g")
	sedFile(sess.extractDir+"/boot/grub/loopback.cfg", "s/Lubuntu/"+sess.cust.DistName+"/g")
	sedFile(sess.extractDir+"/boot/grub/loopback.cfg", "s/Lubuntu/"+sess.cust.DistName+"/g")
	os.Remove(sess.extractDir + "/.disk/release_notes_url")
	writeToFile(sess.extractDir+"/.disk/info", sess.cust.DistName+" "+sess.cust.DistVer+" - Release amd64 ("+date+")")

	writeToFile(sess.chrootDir+"/etc/issue", sess.cust.DistName+" \\r (\\n) (\\l)\n\n")
	writeToFile(sess.chrootDir+"/etc/issue.net", sess.cust.DistName+" "+sess.cust.DistVer)
	writeToFile(sess.chrootDir+"/etc/legal", "")
	writeToFile(sess.chrootDir+"/etc/lsb-release", `DISTRIB_ID=Ubuntu
DISTRIB_RELEASE=`+sess.cust.DistVer+`
DISTRIB_CODENAME=focal
DISTRIB_DESCRIPTION=\"`+sess.cust.DistName+" "+sess.cust.DistVer+`\"
`)

	writeToFile(sess.chrootDir+"/etc/os-release",
		`NAME="`+sess.cust.DistName+`"
VERSION="`+sess.cust.DistVer+`"
ID=ubuntu
ID_LIKE=debian
PRETTY_NAME="`+sess.cust.DistName+" "+sess.cust.DistVer+`"
VERSION_ID="`+sess.cust.DistVer+`"
HOME_URL="https://humaidq.ae"
VERSION_CODENAME=focal
UBUNTU_CODENAME=focal`)

	sedFile(sess.chrootDir+"/etc/calamares/branding/lubuntu/branding.desc",
		"s/Lubuntu 20.04 LTS/"+sess.cust.DistName+" "+sess.cust.DistVer+"/g")
	sedFile(sess.chrootDir+"/etc/calamares/branding/lubuntu/branding.desc",
		"s/Lubuntu/"+sess.cust.DistName+"/g")
	sedFile(sess.chrootDir+"/etc/calamares/branding/lubuntu/branding.desc",
		"s/20.04 LTS/"+sess.cust.DistVer+"/g")
	sedFile(sess.chrootDir+"/etc/calamares/branding/lubuntu/branding.desc",
		"s/lubuntu\\.me/humaidq\\.ae/g")

	// Install packages
	setProgress(sess, "Customising: Creating build script", 7)
	if err := buildCustomiseScript(sess); err != nil {
		return errors.Wrap(err, "build customise script")
	}

	// fix dns resolution issue
	os.Remove(sess.chrootDir + "/etc/resolv.conf")
	os.Remove(sess.chrootDir + "/var/lib/dpkg/statoverride")
	writeToFile(sess.chrootDir+"/etc/resolv.conf", "nameserver 8.8.8.8")
	setProgress(sess, "Customising: Mouting partitions", 7)

	// mount /dev
	_, err := execc(sess.tempDir, "mount", "--bind", "/dev/", sess.chrootDir+"/dev")
	if err != nil {
		return errors.Wrap(err, "mount /dev for chroot")
	}

	setProgress(sess, "Customising: Installing packages and applying customisations", 8)
	if _, err := execc(sess.tempDir, "chroot", sess.chrootDir, "/bin/bash", "/root/cust.sh"); err != nil {
		return errors.Wrap(err, "chroot customise script")
	}

	// copy over pictures
	wd, _ := os.Getwd()
	execc(wd, "cp", "-f", "./assets/splash.pcx", sess.extractDir+"/isolinux/")
	execc(wd, "cp", "-f", "./assets/splash.png", sess.extractDir+"/isolinux/")
	execc(wd, "cp", "-f", "./assets/lubuntu-calamares.desktop", sess.chrootDir+"/usr/share/applications/lubuntu-calamares.desktop")
	execc(wd, "chmod", "+x", sess.chrootDir+"/usr/share/applications/lubuntu-calamares.desktop")
	execc(wd, "cp", "-f", "./assets/04081_lagoonnebula_1920x1080.png", sess.chrootDir+"/usr/share/lubuntu/wallpapers/lubuntu-default-wallpaper.png")

	if sess.cust.StringencyLevel >= 2 {
		sedFile(sess.chrootDir+"/etc/adduser.conf", `/^DIR_MODE=/ s/=[0-9]*\+/=0700/`)
		sedFile(sess.chrootDir+"/etc/login.defs", `/^UMASK\s\+/ s/022/077/`)
		mkdir(sess.chrootDir + "/etc/skel/.mozilla")
		mkdir(sess.chrootDir + "/etc/skel/.mozilla/firefox")
		execc(wd, "cp", "-r", "./assets/qsebcm8n.os", sess.chrootDir+"/etc/skel/.mozilla/firefox/")
		execc(wd, "cp", "-r", "./assets/installs.ini", sess.chrootDir+"/etc/skel/.mozilla/firefox/")
		execc(wd, "cp", "-r", "./assets/profiles.ini", sess.chrootDir+"/etc/skel/.mozilla/firefox/")
	} else if sess.cust.StringencyLevel >= 3 {
		sedFile(sess.chrootDir+"/etc/login.defs", `/^UMASK\s\+/ s/022/077/`)
	}

	umount(sess.chrootDir + "/dev")

	return nil
}

func buildCustomiseScript(sess *buildSession) error {
	var sh strings.Builder
	// Important note: We need to remember to add a new line after every write!

	sh.WriteString(`#!/bin/bash
export HOME=/root
export LC_ALL=C

mount -t proc none /proc
mount -t sysfs none /sys
mount -t devpts none /dev/pts

apt update
#apt upgrade -y --allow-downgrades

`)

	if sess.cust.StringencyLevel >= 1 {
		sess.cust.AddPackages = append(sess.cust.AddPackages, "ufw")
		sess.cust.AddPackages = append(sess.cust.AddPackages, "libpam-cracklib")
		sess.cust.AddPackages = append(sess.cust.AddPackages, "libpam-pwquality")
	}
	if sess.cust.StringencyLevel >= 3 {
		sess.cust.AddPackages = append(sess.cust.AddPackages, "usbguard")
	}
	for _, pkg := range sess.cust.AddPackages {
		sh.WriteString("apt install -y " + pkg + "\n")
	}

	if sess.cust.StringencyLevel >= 1 {
		sh.WriteString("ufw enable\n")
	}

	if sess.cust.StringencyLevel >= 2 {
		sh.WriteString(`echo "TMOUT=\"\$(( 60*10 ))\";\n[ -z \"$DISPLAY\" ] && export TMOUT;\ncase \$( /usr/bin/tty ) in\n	/dev/tty[0-9]*) export TMOUT;;\nesac" | tee -a /etc/skel/.bashrc`)
	}

	sess.cust.RemovePackages = append(sess.cust.RemovePackages, "whoopsie")
	sess.cust.RemovePackages = append(sess.cust.RemovePackages, "apport")
	sess.cust.RemovePackages = append(sess.cust.RemovePackages, "apport-symptoms")
	sess.cust.RemovePackages = append(sess.cust.RemovePackages, "popularity-contest")
	for _, pkg := range sess.cust.RemovePackages {
		sh.WriteString("apt purge -y " + pkg + "\n")
	}

	sh.WriteString("apt autoremove --purge -y\n")

	sh.WriteString(sess.cust.Script + "\n")

	sh.WriteString(`umount /proc
umount /sys
umount /dev/pts
`)

	err := writeToFile(sess.chrootDir+"/root/cust.sh", sh.String())
	if err != nil {
		return err
	}

	return nil
}
