package client

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

// [defaults]
// profile=cog.mycompany.com

// [cog.mycompany.com]
// url=http://cog.mycompany.com:4000
// password=BgEocTFGzON$U39srRt5^fi(ZD0KxV*1
// user=admin

func TestProfileHumanEyeball(t *testing.T) {
	defaultProfile := "cog.mycompany.com"

	profile := Profile{Profiles: make(map[string]ProfileEntry)}
	profile.Defaults.Profile = defaultProfile
	profile.Profiles[defaultProfile] = ProfileEntry{
		URLString: "http://cog.mycompany.com:4000",
		Password:  "BgEocTFGzON$U39srRt5^fi(ZD0KxV*1",
		Username:  "admin",
	}
	profile.Profiles["profile2"] = ProfileEntry{
		URLString: "http://cog.myothercompany.com:4000",
		Password:  "SomeWackyPassword!11!1",
		Username:  "cog-admin",
	}

	y, err := yaml.Marshal(profile)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(y))
}
