package container

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

var (
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
