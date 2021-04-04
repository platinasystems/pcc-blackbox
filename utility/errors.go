package utility

import (
	"fmt"
)

type Errors []error

func (errs Errors) String() (s string) {
	for _, err := range errs {
		s += fmt.Sprintf("%v\n", err)
	}
	return
}

func (errs *Errors) Append(es ...error) {
	for _, e := range es {
		if e != nil {
			*errs = append(*errs, e)
		}
	}
}

func (errs *Errors) Prepend(es ...error) {
	*errs = append(es, *errs...)
}

func (errs *Errors) Empty() {
	*errs = (*errs)[:0]
}

func (errs Errors) IsEmpty() bool {
	return len(errs) == 0
}

func (errs Errors) ToStringSlice() []string {
	s := []string{}
	for _, e := range errs {
		s = append(s, fmt.Sprintf("%v", e))
	}
	return s
}
