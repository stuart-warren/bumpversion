package deps

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"regexp"
	"strings"
)

var (
	from = []byte("FROM")
	// comment           = []byte("#")
	space             = []byte(" ")
	newLine           = []byte("\n")
	dockerImageRegexp = regexp.MustCompile("^(((?P<registry>[a-z]+[.][a-z0-9.-]+(:[0-9]+))/)?((?P<group>[a-zA-Z0-9-]+)/)?(?P<image>[a-z0-9]+[a-z0-9-]+))((:(?P<tag>[a-z0-9._+-]+))?)((@(?P<digest>sha256:[a-f0-9]+))?)$")
)

// DockerImage implements Versioned interface
type DockerImage struct {
	name   string
	tag    string
	digest string
}

func minInt(values ...int) int {
	m := math.MaxInt8
	for _, v := range values {
		if v >= 0 && v < m {
			m = v
		}
	}
	return m
}

// NewDockerImage create an instance of DockerImage
// The name of the image is validated and parsed.
func NewDockerImage(image string) (*DockerImage, error) {
	matches := dockerImageRegexp.MatchString(image)
	if !matches {
		return &DockerImage{}, fmt.Errorf("%s is not a valid image", image)
	}
	var name, digest, tag string
	at := strings.LastIndexByte(image, '@')
	if at >= 0 {
		digest = image[at+1 : len(image)]
	}
	colon := strings.LastIndexByte(image[:minInt(at, len(image))], ':')
	if colon >= 0 {
		tag = image[colon+1 : minInt(at, len(image))]
	}
	name = image[:minInt(at, colon, len(image))]
	return &DockerImage{name: name, tag: tag, digest: digest}, nil
}

// String is used to stringify the DockerImage
func (di DockerImage) String() string {
	if di.tag != "" && di.digest != "" {
		return fmt.Sprintf("%s:%s@%s", di.name, di.tag, di.digest)
	} else if di.digest != "" {
		return fmt.Sprintf("%s@%s", di.name, di.digest)
	} else if di.tag != "" {
		return fmt.Sprintf("%s:%s", di.name, di.tag)
	}
	return di.name
}

// Name returns the name of the DockerImage excluding any version infomation
func (di DockerImage) Name() string {
	return di.name
}

// SetVersion sets the digest (if prefixed with 'sha256:') or tag of the DockerImage
func (di *DockerImage) SetVersion(ver string) {
	if strings.HasPrefix(ver, "sha256:") {
		di.digest = ver
		di.tag = ""
	} else if strings.Contains(ver, "@sha256:") {
		parts := strings.SplitN(ver, "@", 2)
		di.tag = parts[0]
		di.digest = parts[1]
	} else {
		di.tag = ver
		di.digest = ""
	}
}

// DockerFile imliments VersionedPackages interface
type DockerFile struct {
	name      string
	data      []byte
	artifacts map[string]Versioned
}

// NewDockerFile accepts a name and a Reader, parses the contents
func NewDockerFile(name string, r io.Reader) (*DockerFile, error) {
	d := &DockerFile{name: name, artifacts: make(map[string]Versioned)}
	err := d.Load(r)
	return d, err
}

// GetArtifacts returns the map of Artifacts
func (d DockerFile) GetArtifacts() map[string]Versioned {
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
			image, err := NewDockerImage(string(lineParts[1]))
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
