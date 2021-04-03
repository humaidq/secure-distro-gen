package builder

import (
	"fmt"
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

	output, err := proot(sess, "apt update")
	if err != nil {
		fmt.Println(output)
		return errors.Wrap(err, "apt update")
	}

	output, err = proot(sess, "apt upgrade")
	if err != nil {
		fmt.Println(output)
		return errors.Wrap(err, "apt upgrade")
	}

	output, err = proot(sess, "apt install "+strings.Join(sess.cust.AddPackages, " "))

	if err != nil {
		fmt.Println(output)
		return errors.Wrap(err, "apt upgrade")
	}
	return nil
}
