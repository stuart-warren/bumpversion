package dockerfile_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stuart-warren/bumpversion/deps/container/dockerfile"
)

var files = map[string]string{
	"fixtures/Dockerfile.1": "ubuntu:14.04",
	"fixtures/Dockerfile.2": "library/alpine@sha256:59384573945873458347593587",
	"fixtures/Dockerfile.3": "dk.tech.example.com:8080/team1/image2",
	"fixtures/Dockerfile.4": "alpine:3.5@sha256:59384573945873458347593587",
}

func TestSetVersion(t *testing.T) {
	f, _ := os.Open("fixtures/Dockerfile.4")
	df, _ := dockerfile.New("alpine:3.5", f)
	df.SetVersion("alpine", "3.6@sha256:9887454752654746548375")
	got := df.GetArtifacts()["alpine"].String()
	expected := "alpine:3.6@sha256:9887454752654746548375"
	if got != expected {
		t.Errorf("got %s expected %s", got, expected)
	}
}

func TestParseFile(t *testing.T) {
	for file, expected := range files {
		t.Run(fmt.Sprintf("with %s", file), func(t *testing.T) {
			f, _ := os.Open(file)
			d, err := dockerfile.New(file, f)
			if err != nil {
				t.Errorf("Failed to read %s as a Dockerfile: %s", file, err)
			}
			if len(d.GetArtifacts()) < 1 {
				t.Errorf("Expected at least one artifact, got %d", len(d.GetArtifacts()))
			} else {
				for _, image := range d.GetArtifacts() {
					if image.String() != expected {
						t.Errorf("Expected %s, got %+v", expected, d.GetArtifacts())
					}
				}
			}
		})
	}
}

func TestParseBadFile(t *testing.T) {
	f, _ := os.Open("fixtures/Dockerfile.bad")
	_, err := dockerfile.New("badimage", f)
	if err == nil {
		t.Errorf("expected error reading bad image from file")
	}
}

func TestNotParseFile(t *testing.T) {
	f, _ := os.Open("fixtures/nonexistant")
	_, err := dockerfile.New("nonexistant", f)
	if err == nil {
		t.Errorf("expected error reading non-existant file")
	}
}

func TestReadWriteFile(t *testing.T) {
	file := "fixtures/Dockerfile.1"
	in := new(bytes.Buffer)
	f, _ := os.Open(file)
	d, err := dockerfile.New(file, io.TeeReader(f, in))
	if err != nil {
		t.Errorf("Failed to read %s as a Dockerfile: %s", file, err)
	}
	out := new(bytes.Buffer)
	d.Write(out)
	if len(in.Bytes()) != len(out.Bytes()) {
		t.Errorf("length of 'in' %d does not match 'out' %d", len(in.Bytes()), len(out.Bytes()))
	}
	if !bytes.Equal(in.Bytes(), out.Bytes()) {
		t.Error("'in' does not match 'out'")
		for i, _ := range in.Bytes() {
			t.Logf("%s %s\n", in.Bytes()[i], out.Bytes()[i])
		}
	}
}

func TestModifyFile(t *testing.T) {
	in := new(bytes.Buffer)
	content := []byte("FROM ubuntu:14.04\nENTRYPOINT [\"/usr/bin/bash\"]")
	in.Write(content)
	d, err := dockerfile.New("somefile", in)
	in.Reset()
	if err != nil {
		t.Errorf("Failed to read Dockerfile: %s", err)
	}
	err = d.SetVersion("ubuntu", "16.04")
	if err != nil {
		t.Errorf("could not find image %q", "ubuntu")
	}
	out := new(bytes.Buffer)
	d.Write(out)
	if bytes.Equal(content, out.Bytes()) {
		t.Error("'in' matches 'out' when it should be different")
		for i, _ := range in.Bytes() {
			t.Logf("%s %s\n", in.Bytes()[i], out.Bytes()[i])
		}
	}
}

func TestModifyFileWithDigest(t *testing.T) {
	in := new(bytes.Buffer)
	content := []byte("FROM ubuntu:14.04\nENTRYPOINT [\"/usr/bin/bash\"]")
	in.Write(content)
	d, err := dockerfile.New("somefile", in)
	in.Reset()
	if err != nil {
		t.Errorf("Failed to read Dockerfile: %s", err)
	}
	err = d.SetVersion("ubuntu", "sha256:987459348576783645")
	if err != nil {
		t.Errorf("could not find image %q", "ubuntu")
	}
	if d.GetArtifacts()["ubuntu"].String() != "ubuntu@sha256:987459348576783645" {
		t.Errorf("version is not what is expected: %s", d.GetArtifacts()["ubuntu"].String())
	}
}

func TestFailToModifyFile(t *testing.T) {
	in := new(bytes.Buffer)
	content := []byte("FROM ubuntu:14.04\nENTRYPOINT [\"/usr/bin/bash\"]\n")
	in.Write(content)
	d, err := dockerfile.New("somefile", in)
	if err != nil {
		t.Errorf("Failed to read Dockerfile: %s", err)
	}
	err = d.SetVersion("nonexistant", "16.04")
	if err == nil {
		t.Errorf("found image %q when it shouldn't have", "nonexistant")
	}
	out := new(bytes.Buffer)
	d.Write(out)
	if len(content) != len(out.Bytes()) {
		t.Errorf("length of 'in' %d does not match 'out' %d", len(content), len(out.Bytes()))
	}
	if !bytes.Equal(content, out.Bytes()) {
		t.Error("'in' does not match 'out' when it should be unchanged")
		for i, _ := range content {
			t.Logf("%s %s\n", content[i], out.Bytes()[i])
		}
	}
}
