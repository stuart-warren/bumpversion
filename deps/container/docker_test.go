package container_test

import (
	"fmt"
	"testing"

	"github.com/stuart-warren/bumpversion/deps/container"
)

var images = []struct {
	input string
	name  string
	err   bool
}{
	{"invalid.image:bad", "", true},
	{"alpine:3.5@sha256:59384573945873458347593587", "alpine", false},
	{"ubuntu", "ubuntu", false},
	{"ubuntu:14.04", "ubuntu", false},
	{"library/alpine:3.5", "library/alpine", false},
	{"library/alpine@sha256:59384573945873458347593587", "library/alpine", false},
	{"dk.tech.example.com:8080/team1/image2:latest", "dk.tech.example.com:8080/team1/image2", false},
}

func TestDockerImageParse(t *testing.T) {
	for _, image := range images {
		di, err := container.NewDockerImage(image.input)
		t.Run(fmt.Sprintf("with %s", image.input), func(t *testing.T) {
			if image.err {
				if err == nil {
					t.Errorf("expected error, didn't get one")
				}
			} else if di.String() != image.input {
				t.Errorf("got %s expected %s", di.String(), image.input)
			}
		})
		t.Run(fmt.Sprintf("with %s", image.input), func(t *testing.T) {
			if di.Name() != image.name {
				t.Errorf("got %s expected %s", di.Name(), image.name)
			}
		})
	}
}

func TestDockerSetVersion(t *testing.T) {
	image := "alpine"
	versions := []struct {
		input    string
		expected string
	}{
		{"3.6", "alpine:3.6"},
		{"sha256:59384573945873458347593587", "alpine@sha256:59384573945873458347593587"},
		{"3.6@sha256:59384573945873458347593587", "alpine:3.6@sha256:59384573945873458347593587"},
		{"", "alpine"},
	}
	for _, ver := range versions {
		di, _ := container.NewDockerImage(image)
		t.Run(fmt.Sprintf("with %s", ver.input), func(t *testing.T) {
			di.SetVersion(ver.input)
			if di.String() != ver.expected {
				t.Errorf("got %s expected %s", di.String(), ver.expected)
			}
		})
	}
}
