package dockerfile

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dc *Client
)

func TestMain(m *testing.M) {
	var err error

	dc, err = New(Options{
		Filename: "./fixtures/valid.Dockerfile",
	})
	if err != nil {
		panic(err.Error())
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	assert.NotNil(t, dc)
}

func TestLoadFile(t *testing.T) {
	cases := []struct {
		name    string
		dfile   string
		wantErr bool
	}{
		{
			name:    "Failed on non-existing file",
			dfile:   "",
			wantErr: true,
		},
		{
			name:    "Fail on empty file",
			dfile:   "./fixtures/empty.Dockerfile",
			wantErr: true,
		},
		{
			name:    "Fail on wrong file format",
			dfile:   "./fixtures/invalid.Dockerfile",
			wantErr: true,
		},
		{
			name:    "Valid",
			dfile:   "./fixtures/valid.Dockerfile",
			wantErr: false,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			c, err := New(Options{
				Filename: tt.dfile,
			})
			if tt.wantErr {
				fmt.Println(err)
				require.Error(t, err)
				return
			}
			assert.NotNil(t, c)
		})
	}
}
