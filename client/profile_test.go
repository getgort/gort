package client

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

// [defaults]
// profile=gort.mycompany.com

// [gort.mycompany.com]
// url=http://gort.mycompany.com:4000
// password=BgEocTFGzON$U39srRt5^fi(ZD0KxV*1
// user=admin

func TestProfileHumanEyeball(t *testing.T) {
	defaultProfile := "gort.mycompany.com"

	profile := Profile{Profiles: make(map[string]ProfileEntry)}
	profile.Defaults.Profile = defaultProfile
	profile.Profiles[defaultProfile] = ProfileEntry{
		URLString: "http://gort.mycompany.com:4000",
		Password:  "BgEocTFGzON$U39srRt5^fi(ZD0KxV*1",
		Username:  "admin",
	}
	profile.Profiles["profile2"] = ProfileEntry{
		URLString: "http://gort.myothercompany.com:4000",
		Password:  "SomeWackyPassword!11!1",
		Username:  "gort-admin",
	}

	y, err := yaml.Marshal(profile)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(y))
}
