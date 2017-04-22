package dockerfile

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/stuart-warren/bumpversion/deps"
	"github.com/stuart-warren/bumpversion/deps/container"
)

var (
	from = []byte("FROM")
	// comment           = []byte("#")
	space   = []byte(" ")
	newLine = []byte("\n")
)

// DockerFile imliments VersionedPackages interface
type DockerFile struct {
	name      string
	data      []byte
	artifacts map[string]deps.Versioned
}

// New accepts a name and a Reader, parses the contents
func New(name string, r io.Reader) (*DockerFile, error) {
	d := &DockerFile{name: name, artifacts: make(map[string]deps.Versioned)}
	err := d.Load(r)
	return d, err
}

// GetArtifacts returns the map of Artifacts
func (d DockerFile) GetArtifacts() map[string]deps.Versioned {
	return d.artifacts
}

// Load accepts a reader and parses the input
func (d *DockerFile) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	data := bytes.NewBuffer(d.data)
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, from) {
			lineParts := bytes.SplitN(line, space, 3)
			image, err := container.NewDockerImage(string(lineParts[1]))
			if err != nil {
				return err
			}
			d.GetArtifacts()[image.Name()] = image
		}
		_, _ = data.Write(append(line, newLine...))
		d.data = data.Bytes()
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// SetVersion allows you to change the version of one of the images if they exist
func (d *DockerFile) SetVersion(image, ver string) error {
	if i, ok := d.GetArtifacts()[image]; ok {
		data := new(bytes.Buffer)
		find := append(from, []byte(" "+i.String())...)
		scanner := bufio.NewScanner(bytes.NewReader(d.data))
		for scanner.Scan() {
			line := scanner.Bytes()
			if bytes.HasPrefix(line, find) {
				i.SetVersion(ver)
				new := append(from, []byte(" "+i.String())...)
				line = new
			}
			_, _ = data.Write(append(line, newLine...))
		}
		d.data = data.Bytes()
	} else {
		return fmt.Errorf("could not find image %q", image)
	}
	return nil
}

// Write is used to write out the DockerFile
func (d DockerFile) Write(w io.Writer) error {
	_, err := w.Write(d.data)
	return err
}
