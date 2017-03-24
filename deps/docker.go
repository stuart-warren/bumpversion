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
	From              = []byte("FROM")
	Comment           = []byte("#")
	Space             = []byte(" ")
	NewLine           = []byte("\n")
	DockerImageRegexp = regexp.MustCompile("^(((?P<registry>[a-z]+[.][a-z0-9.-]+(:[0-9]+))/)?((?P<group>[a-zA-Z0-9-]+)/)?(?P<image>[a-z0-9]+[a-z0-9-]+))((:(?P<tag>[a-z0-9._+-]+))?)((@(?P<digest>sha256:[a-f0-9]+))?)$")
)

type dockerImage struct {
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

func NewDockerImage(image string) (*dockerImage, error) {
	matches := DockerImageRegexp.MatchString(image)
	if !matches {
		return &dockerImage{}, fmt.Errorf("%s is not a valid image", image)
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
	return &dockerImage{name: name, tag: tag, digest: digest}, nil
}

func (di dockerImage) String() string {
	if di.tag != "" && di.digest != "" {
		return fmt.Sprintf("%s:%s@%s", di.name, di.tag, di.digest)
	} else if di.digest != "" {
		return fmt.Sprintf("%s@%s", di.name, di.digest)
	} else if di.tag != "" {
		return fmt.Sprintf("%s:%s", di.name, di.tag)
	}
	return di.name
}

func (di dockerImage) Name() string {
	return di.name
}

func (di *dockerImage) SetVersion(ver string) {
	if strings.HasPrefix(ver, "sha256:") {
		di.digest = ver
		di.tag = ""
	} else {
		di.tag = ver
		di.digest = ""
	}
}

type dockerFile struct {
	name      string
	data      []byte
	Artifacts map[string]Versioned
}

// NewDockerFile accepts a name and a Reader, parses the contents
func NewDockerFile(name string, r io.Reader) (*dockerFile, error) {
	d := &dockerFile{name: name, Artifacts: make(map[string]Versioned)}
	err := d.Load(r)
	return d, err
}

func (d *dockerFile) Load(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	data := bytes.NewBuffer(d.data)
	for scanner.Scan() {
		line := scanner.Bytes()
		if bytes.HasPrefix(line, From) {
			lineParts := bytes.SplitN(line, Space, 3)
			image, err := NewDockerImage(string(lineParts[1]))
			if err != nil {
				return err
			}
			d.Artifacts[image.Name()] = image
		}
		_, _ = data.Write(append(line, NewLine...))
		d.data = data.Bytes()
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

func (d *dockerFile) SetVersion(image, ver string) error {
	if i, ok := d.Artifacts[image]; ok {
		data := new(bytes.Buffer)
		find := append(From, []byte(" "+i.String())...)
		scanner := bufio.NewScanner(bytes.NewReader(d.data))
		for scanner.Scan() {
			line := scanner.Bytes()
			if bytes.HasPrefix(line, find) {
				i.SetVersion(ver)
				new := append(From, []byte(" "+i.String())...)
				line = new
			}
			_, _ = data.Write(append(line, NewLine...))
		}
		d.data = data.Bytes()
	} else {
		return fmt.Errorf("could not find image %q", image)
	}
	return nil
}

func (d dockerFile) Write(w io.Writer) error {
	_, err := w.Write(d.data)
	return err
}
