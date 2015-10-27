package cfg

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu     sync.Mutex
	config *Cfg
)

type (
	Cfg struct {
		AWSAccessKeyID string `json:"aws_access_key_id"`
		AWSSecretKey   string `json:"aws_secret_key"`
		Region         string `json:"region"`
	}

	InArgs struct {
		Bucket   string
		Prefix   string
		LocalDir string
		Regexp   string
		DryRun   bool
	}
)

func GetCfg() *Cfg {
	if nil == config {
		mu.Lock()
		defer mu.Unlock()

		err := loadCfg()
		if err != nil {
			fmt.Println(">", err)
			os.Exit(0)
		}
	}

	return config
}

func loadCfg() error {
	errfmt := "fs.loadCfg error: %s"

	b, err := load("config.json")
	if err != nil {
		return fmt.Errorf(errfmt, err)
	}

	var cfg Cfg
	if err := json.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf(errfmt, err)
	}

	config = &cfg
	return nil
}

func load(fname string) ([]byte, error) {
	dir, err := filepath.Abs(filepath.Dir("./"))
	if err != nil {
		return nil, err
	}

	fpath := filepath.Join(dir, fname)

	b, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	return b, nil
}
