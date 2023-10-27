// package adapted from upstream: https://github.com/sachaos/todoist
package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rkoesters/xdg/basedir"
	"github.com/spf13/viper"

	todoist "github.com/sachaos/todoist/lib"
)

const (
	configName = "config"
	configType = "json"
)

var (
	cachePath           = filepath.Join(basedir.CacheHome, "tuidoist", "cache.json")
	opsPath             = filepath.Join(basedir.CacheHome, "tuidoist", "ops.json")
	configPath          = filepath.Join(basedir.ConfigHome, "tuidoist")
	ShortDateTimeFormat = "06/01/02(Mon) 15:04"
	ShortDateFormat     = "06/01/02(Mon)"
)

func Exists(path string) (bool, error) {
	_, fileErr := os.Stat(path)
	if fileErr == nil {
		return true, nil
	}
	if os.IsNotExist(fileErr) {
		return false, nil
	}
	return true, nil
}

func AssureExists(filePath string) error {
	path := filepath.Dir(filePath)
	exists, err := Exists(path)
	if err != nil {
		return err
	}
	if !exists {
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Couldn't create path: %s", path)
		}
	}
	return nil
}

func LoadCache(s *todoist.Store, l *todoist.Store, cmds *todoist.Commands) error {
	err := ReadCache(s, l, cmds)
	if err != nil {
		err = WriteCache(s, cmds)
		if err != nil {
			return err
		}
	}
	return nil
}

func ReadCache(s *todoist.Store, l *todoist.Store, o *todoist.Commands) error {
	jsonString, err := os.ReadFile(cachePath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonString, &s)

	jsonOpsString, err := os.ReadFile(opsPath)
	if err != nil {
		return err
	}
	json.Unmarshal(jsonString, &l)

	var ops []todoist.Command
	err = json.Unmarshal(jsonOpsString, &ops)
	if err != nil {
		return err
	}
	*o = todoist.Commands(ops)
	s.ConstructItemTree()
	l.ConstructItemTree()
	return nil
}

func WriteCache(s *todoist.Store, o *todoist.Commands) error {
	buf, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	buf2, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	err = AssureExists(cachePath)
	if err != nil {
		return err
	}
	err2 := os.WriteFile(cachePath, buf, os.ModePerm)
	if err2 != nil {
		return errors.New("Couldn't write to the cache file")
	}
	err3 := os.WriteFile(opsPath, buf2, os.ModePerm)
	if err3 != nil {
		return errors.New("Couldn't write to the ops file")
	}
	return nil
}

func GetClient(logger func(...any)) (*todoist.Client, *todoist.Store, *todoist.Commands) {
	var store todoist.Store
	var local todoist.Store
	var ops todoist.Commands

	if err := LoadCache(&store, &local, &ops); err != nil {
		panic(err)
	}

	viper.SetConfigType(configType)
	viper.SetConfigName(configName)
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("tuidoist")
	viper.AutomaticEnv()

	var token string

	configFile := filepath.Join(configPath, configName+"."+configType)
	if err := AssureExists(configFile); err != nil {
		panic(err)
	}

	if err := AssureExists(opsPath); err != nil {
		panic(err)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, isConfigNotFoundError := err.(viper.ConfigFileNotFoundError); !isConfigNotFoundError {
			// config file was found but could not be read => not recoverable
			panic(err)
		} else if !viper.IsSet("token") {
			// config file not found and token missing (not provided via another source,
			// such as environment variables) => ask interactively for token and store it in config file.
			fmt.Printf("Input API Token: ")
			fmt.Scan(&token)
			viper.Set("token", token)
			buf, err := json.MarshalIndent(viper.AllSettings(), "", "  ")
			if err != nil {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
			err = os.WriteFile(configFile, buf, 0600)
			if err != nil {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
		}
	}

	if exists, _ := Exists(configFile); exists {
		// Ensure that the config file has permission 0600, because it contains
		// the API token and should only be read by the user.
		// This is only necessary iff the config file exists, which may not be the case
		// when config is loaded from environment variables.
		fi, err := os.Lstat(configFile)
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
		if runtime.GOOS != "windows" && fi.Mode().Perm() != 0600 {
			panic(fmt.Errorf("Config file has wrong permissions. Make sure to give permissions 600 to file %s \n", configFile))
		}
	}
	config := &todoist.Config{
		AccessToken:    viper.GetString("token"),
		Color:          viper.GetBool("color"),
		DateFormat:     viper.GetString("shortdateformat"),
		DateTimeFormat: viper.GetString("shortdatetimeformat"),
		DebugMode:      false, // len(os.Getenv("DEBUG")) > 0,
	}

	client := todoist.NewClient(config)
	client.Store = &store
	if len(store.Projects) == 0 {
		err := client.Sync(context.Background())
		if err != nil {
			logger("Sync err", err)
		}
		WriteCache(&store, &ops)
	}
	return client, &local, &ops
}
