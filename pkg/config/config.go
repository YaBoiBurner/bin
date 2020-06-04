package config

import (
	"encoding/json"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/apex/log"
)

var cfg config

type config struct {
	DefaultPath string             `json:"default_path"`
	Bins        map[string]*Binary `json:"bins"`
}

type Binary struct {
	Path       string `json:"path"`
	RemoteName string `json:"remote_name"`
	Version    string `json:"version"`
	Hash       string `json:"hash"`
	URL        string `json:"url"`
}

func CheckAndLoad() error {
	u, _ := user.Current()
	f, err := os.OpenFile(filepath.Join(u.HomeDir, ".bin/config.json"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	defer f.Close()

	err = json.NewDecoder(f).Decode(&cfg)

	if len(cfg.DefaultPath) == 0 {
		cfg.DefaultPath = getDefaultPath()
	}
	log.Debugf("Download path set to %s", cfg.DefaultPath)
	// ignore if file is empty
	if err != nil && err != io.EOF {
		return err
	} else if err == io.EOF {
		cfg.Bins = map[string]*Binary{}
	}

	return nil

}

//getDefaultPath reads the user's PATH variable
//and returns the first directory that's writable by the current
//user in the system
func getDefaultPath() string {
	penv := os.Getenv("PATH")
	log.Debugf("User PATH is [%s]", penv)
	for _, p := range strings.Split(penv, ":") {
		log.Debugf("Checking path %s", p)

		fi, _ := os.Stat(p)
		// If it's a dir and has the write bit set
		if fi.IsDir() && !(fi.Mode().Perm()&(1<<(uint(7))) == 0) {
			log.Debugf("%s seems to be a dir and writable, using it.", p)
			return p
		}

	}

	return ""

}

func Get() *config {
	return &cfg
}

//UpsertBinary adds or updats an existing
//binary resource in the config
func UpsertBinary(c *Binary) error {

	if c != nil {
		cfg.Bins[c.Path] = c
		err := write()
		if err != nil {
			return err
		}
	}

	return nil
}

// RemoveBinaries removes the specified paths
// from bin configuration. It doesn't care about the order
func RemoveBinaries(paths []string) error {
	for _, p := range paths {
		delete(cfg.Bins, p)
	}

	return write()
}

func write() error {
	u, _ := user.Current()
	f, err := os.OpenFile(filepath.Join(u.HomeDir, ".bin/config.json"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}

	err = json.NewEncoder(f).Encode(cfg)

	if err != nil {
		return err
	}

	return nil
}

// GetArch is the running program's operating system target:
// one of darwin, freebsd, linux, and so on.
func GetArch() []string {
	res := []string{runtime.GOARCH}
	if runtime.GOARCH == "amd64" {
		//Adding x86_64 manually since the uname syscall (man 2 uname)
		//is not implemented in all systems
		res = append(res, "x86_64")
	}
	return res
}

// GetOS is the running program's architecture target:
// one of 386, amd64, arm, s390x, and so on.
func GetOS() []string {
	return []string{runtime.GOOS}
}
