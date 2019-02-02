package bundle

import (
	"encoding/json"
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestInputOutputHumanEyeball(t *testing.T) {
	bundle, err := loadBundle("../testing/test-bundle.yml")
	if err != nil {
		t.Error(err.Error())
	}

	y, err := yaml.Marshal(bundle)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(y))
}
func TestInputOutputHumanEyeballJSON(t *testing.T) {
	bundle, err := loadBundle("../testing/test-bundle.yml")
	if err != nil {
		t.Error(err.Error())
	}

	j, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return
	}

	fmt.Println(string(j))
}
