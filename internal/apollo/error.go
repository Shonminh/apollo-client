package apollo

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

func ListFormatFunc(es []error) string {
	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred:%s", es[0])
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("err %d- %s", i, err)
	}

	return fmt.Sprintf(
		"%d errors occurred: %s",
		len(es), strings.Join(points, "|||"))
}

func NewMutliError() *multierror.Error {
	return &multierror.Error{ErrorFormat: ListFormatFunc}
}

var ErrInvalidHttpStatus = errors.New("invalid http status")
