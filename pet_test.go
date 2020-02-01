package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadConfig(t *testing.T) {

	tcs := []struct {
		name        string
		input       string
		environment map[string]string
		want        []Pet
	}{
		{
			name:  "basic",
			input: "testdata/basic.hcl",
			want: []Pet{
				&Cat{Name: "Ink", Sound: "meow"},
				&Dog{Name: "Swinney", Breed: "Dachshund"},
			},
		},
		{
			name:  "variables",
			input: "testdata/variables.hcl",
			environment: map[string]string{
				"CAT_SOUND": "nyan",
			},
			want: []Pet{
				&Cat{Name: "Neko", Sound: "nyan"},
				&Cat{Name: "Whiskers", Sound: "meow"},
			},
		},
		{
			name:  "functions",
			input: "testdata/function.hcl",
			want: []Pet{
				&Dog{Name: "Spot", Breed: "Pug"},
			},
		},
	}

	for _, tc := range tcs {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.environment {
				os.Setenv(k, v)
			}

			got, err := ReadConfig(tc.input)
			if assert.Nil(t, err, "error while parsing input") {
				assert.Equal(t, tc.want, got)
			} else {
				assert.Fail(t, err.Error())
			}
		})
	}
}
