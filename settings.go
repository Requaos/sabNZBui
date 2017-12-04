package main

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/spf13/viper"
)

func getSettings() map[string]string {
	// example file: secrets.toml
	// [settings]
	// nzbSite = "https://api.nzbplanet.net"
	// nzbKey = "157b4974da310d1f834644fe93298171"
	// sabSite = "localhost:8080"
	// sabKey = "6a1c4e43be73e58c2c2617043c72b8de"
	viper.SetConfigName("secrets")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Config file not found...%s", err.Error())
	}
	settingsFile := viper.GetStringMapString("settings")
	fmt.Println(len(settingsFile))
	if len(settingsFile) > 0 {
		err = settingsToDB(settingsFile)
		if err != nil {
			fmt.Printf("DB Update error...%s", err.Error())
		}
	}
	settings := settingsFromDB()
	fmt.Println("Returning from getSettings()")
	return settings
}

func setSettings(nzbSite string, nzbKey string, sabSite string, sabKey string) {
	settings := map[string]string{
		"nzbsite": nzbSite,
		"nzbkey":  nzbKey,
		"sabsite": sabSite,
		"sabkey":  sabKey,
	}
	settingsToDB(settings)
	startingUp = true
	Settings = settings
	SABnzbd = SABnzbdSession()
}

func settingsFromDB() map[string]string {
	m := make(map[string]string)
	db, err := bolt.Open("settings.db", 0644, nil)
	if err != nil {
		fmt.Printf("Error opening DB...%s\n", err.Error())
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("settings"))
		for _, x := range []string{"nzbsite", "nzbkey", "sabsite", "sabkey"} {
			v := b.Get([]byte(x))
			m[x] = string(v)
			fmt.Println("Retrieving -> " + x + ": " + string(v))
		}
		return nil
	})
	return m
}

func settingsToDB(m map[string]string) error {
	fmt.Println(m)
	db, err := bolt.Open("settings.db", 0644, nil)
	if err != nil {
		fmt.Printf("Error opening DB...%s\n", err.Error())
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		fmt.Println("inside DB update function")
		b, err := tx.CreateBucketIfNotExists([]byte("settings"))
		if err != nil {
			return fmt.Errorf("Bucket error: %s", err.Error())
		}
		for k, v := range m {
			err := b.Put([]byte(k), []byte(v))
			fmt.Println("Inserting --> " + k + ": " + v)
			if err != nil {
				return fmt.Errorf("Error inserting key/value pair into DB bucket...%s", err.Error())
			}
		}
		fmt.Println("exiting DB update function")
		return nil
	})
	fmt.Println("exiting settingsToDB function")
	return err
}
