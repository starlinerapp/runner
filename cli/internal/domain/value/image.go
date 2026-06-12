package value

import (
	"fmt"
	"strings"
)

type ImageRef struct {
	host       string
	repository string
	tag        string
}

func ParseImageRef(s string) (ImageRef, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return ImageRef{}, fmt.Errorf("image reference is required")
	}
	s = strings.TrimSuffix(s, "/")

	var tag string
	if i := strings.LastIndex(s, ":"); i != -1 {
		if strings.LastIndex(s, "/") < i {
			tag = s[i+1:]
			s = s[:i]
			if tag == "" {
				return ImageRef{}, fmt.Errorf("image tag is empty")
			}
		}
	}

	var host, repository string
	if slash := strings.Index(s, "/"); slash != -1 {
		candidate := s[:slash]
		rest := s[slash+1:]
		if strings.ContainsAny(candidate, ".:") || candidate == "localhost" {
			host = candidate
			repository = rest
		} else {
			repository = s
		}
	} else {
		repository = s
	}

	if repository == "" {
		return ImageRef{}, fmt.Errorf("image repository is required")
	}

	return ImageRef{
		host:       host,
		repository: repository,
		tag:        tag,
	}, nil
}

func (r ImageRef) WithTag(tag string) ImageRef {
	if r.tag != "" || tag == "" {
		return r
	}
	r.tag = tag
	return r
}

func (r ImageRef) String() string {
	name := r.repository
	if r.host != "" {
		name = r.host + "/" + r.repository
	}
	if r.tag != "" {
		name += ":" + r.tag
	}
	return name
}

func (r ImageRef) RegistryHost() string {
	return r.host
}
