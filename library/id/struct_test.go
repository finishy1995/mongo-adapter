package id

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWithID_InitID(t *testing.T) {
	s := new(WithID)
	s.InitID()
	assert.NotEmpty(t, s.id, "InitID not run")
}

func TestWithID_GetID(t *testing.T) {
	r := require.New(t)
	s := new(WithID)
	s.InitID()
	r.NotEmpty(s.GetID(), "GetID return wrong")
	s2 := new(WithID)
	s2.InitID()
	r.NotEqual(s.GetID(), s2.GetID(), "Return same id")
}
