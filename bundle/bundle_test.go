package bundle

import (
	"fmt"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestInputOutputHumanEyeball(t *testing.T) {
	bundle, err := loadBundle("../test-bundle.yml")
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
