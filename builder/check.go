package builder

import (
	"errors"
	"os/exec"
	"strings"
)

func DependencyCheck() error {
	var missing []string

	// checks go here
	if _, has := hasCommand("rsync", "-V"); !has {
		missing = append(missing, "rsync")
	}
	if _, has := hasCommand("proot", "-V"); !has {
		missing = append(missing, "PRoot")
	}
	if _, has := hasCommand("mksquashfs", "--help"); !has {
		missing = append(missing, "Squashfs-Tools")
	}
	if _, has := hasCommand("xorriso", "--version"); !has {
		missing = append(missing, "GNU xorriso")
	}

	// when no missing libraries exist
	if len(missing) == 0 {
		return nil
	}

	var s strings.Builder
	s.WriteString("missing dependencies, please install: ")
	for i, v := range missing {
		s.WriteString(v)
		if i < len(missing)-1 {
			s.WriteString(", ")
		}
	}
	return errors.New(s.String())
}

func hasCommand(cmd string, arg ...string) (string, bool) {
	c := exec.Command(cmd, arg...)
	o, err := c.Output()
	return string(o), (err == nil)
}
