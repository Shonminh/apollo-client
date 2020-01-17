package apollo

import "strings"

type Resolver interface {
	resolve() ([]string, error)
}

type singleHostResolver struct {
	host string
}

func NewSingleHostResolver(host string) Resolver {
	return &singleHostResolver{host: host}
}

func (s *singleHostResolver) resolve() ([]string, error) {
	ret := make([]string, 1)
	if strings.HasPrefix(s.host, "http") {
		ret[0] = s.host
	} else {
		ret[0] = "http://" + s.host
	}
	return ret, nil
}
