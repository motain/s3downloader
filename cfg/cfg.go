package cfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	once   sync.Once
	config *Cfg

	ErrUndefinedBucket = errors.New("bucket is not defined")
)

type (
	Cfg struct {
		AWSAccessKeyID string `json:"aws_access_key_id"`
		AWSSecretKey   string `json:"aws_secret_key"`
		Region         string `json:"region"`
	}

	InArgs struct {
		Bucket      string
		Prefix      string
		LocalDir    string
		Regexp      string
		DryRun      bool
		PrependName bool
	}
)

func GetCfg() (*Cfg, error) {
	var err error
	once.Do(func() {
		err = loadCfg()
	})

	if err != nil {
		return nil, fmt.Errorf("cfg.GetCfg: %s", err)
	}

	return config, nil
}

func (in *InArgs) Validate() error {
	if in.Bucket == "" {
		return ErrUndefinedBucket
	}

	return nil
}

func loadCfg() error {
	b, err := load("config.json")
	if err != nil {
		return err
	}

	var cfg Cfg
	if err := json.Unmarshal(b, &cfg); err != nil {
		return err
	}

	config = &cfg
	return nil
}

func load(fname string) ([]byte, error) {
	_, callerName, _, _ := runtime.Caller(0)

	dir, err := filepath.Abs(filepath.Join(filepath.Dir(callerName), "/../"))
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
