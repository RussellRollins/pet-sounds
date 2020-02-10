package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

const (
	environmentKey = "env"
	catSoundKey    = "CAT_SOUND"

	defaultCatSound = "meow"
	defaultDogBreed = "mutt"
)

// The Pet interface is used to implement the "application" logic of our toy
// example here. Each Pet is represented in hcl as:
//   pet "<PET NAME>" {
//     type = "<dog | cat>"
//     characteristics {
//       // characteristics unique to dogs or cats
//     }
//   }
type Pet interface {
	Say()
	Act()
}

// PetsHCL is a generic structure that could be either cats or dogs. The Type
// field indicates which, and the generic "characteristics" block HCL will be
// decoded into the unique fields for that type.
// Note the use of the `hcl:",remain"` tag, which puts all undecoded HCL into
// an hcl.Body for use later.
type PetsHCL struct {
	PetHCLBodies []*struct {
		Name               string `hcl:",label"`
		Type               string `hcl:"type"`
		CharacteristicsHCL *struct {
			HCL hcl.Body `hcl:",remain"`
		} `hcl:"characteristics,block"`
	} `hcl:"pet,block"`
}

// Note the optional `hcl:"sound,optional"` tag on the Sound field. This Field
// is unique to cats, and a dog characteristic block would have a type error
// when decoding.
type Cat struct {
	Name  string
	Sound string `hcl:"sound,optional"`
}

// Implement the Pet interface.
func (c *Cat) Say() {
	fmt.Printf("%s %s\n", c.Name, c.Sound)
}
func (c *Cat) Act() {
	fmt.Printf("%s snoozes\n", c.Name)
}

// Note the optional `hcl:"breed,optional"` tag on the Breed field. This Field
// is unique to dogs, and a cat characteristic block would have a type error
// when decoding.
type Dog struct {
	Name  string
	Breed string `hcl:"breed,optional"`
}

// Implement the Pet interface.
func (d *Dog) Say() {
	fmt.Printf("%s the %s barks\n", d.Name, d.Breed)
}
func (d *Dog) Act() {
	fmt.Printf("%s the %s plays\n", d.Name, d.Breed)
}

// ReadConfig decodes the HCL file at filename into a slice of Pets and returns
// it.
func ReadConfig(filename string) ([]Pet, error) {
	// First, open a file handle to the input filename.
	input, err := os.Open(filename)
	if err != nil {
		return []Pet{}, fmt.Errorf(
			"error in ReadConfig openin pet config file: %w", err,
		)
	}
	defer input.Close()

	// Next, read that file into a byte slice for use as a buffer. Because HCL
	// decoding must happen in the context of a whole file, it does not take an
	// io.Reader as an input, instead relying on byte slices.
	src, err := ioutil.ReadAll(input)
	if err != nil {
		return []Pet{}, fmt.Errorf(
			"error in ReadConfig reading input `%s`: %w", filename, err,
		)
	}

	// Instantiate an HCL parser with the source byte slice.
	parser := hclparse.NewParser()
	srcHCL, diag := parser.ParseHCL(src, filename)
	if diag.HasErrors() {
		return []Pet{}, fmt.Errorf(
			"error in ReadConfig parsing HCL: %w", diag,
		)
	}

	// Call a helper function which creates an HCL context for use in
	// decoding the parsed HCL.
	evalContext, err := createContext()
	if err != nil {
		return []Pet{}, fmt.Errorf(
			"error in ReadConfig creating HCL evaluation context: %w", err,
		)
	}

	// Start the first pass of decoding. This decodes all pet blocks into
	// a generic form, with a Type field for use in determining whether they
	// are cats or dogs. The configuration in the characteristics will be left
	// undecoded in an hcl.Body. This Body will be decoded into different pet
	// types later, once the context of the Type is known.
	petsHCL := &PetsHCL{}
	if diag := gohcl.DecodeBody(srcHCL.Body, evalContext, petsHCL); diag.HasErrors() {
		return []Pet{}, fmt.Errorf(
			"error in ReadConfig decoding HCL configuration: %w", diag,
		)
	}

	// Iterate through the generic pets, switch on type, then decode the
	// hcl.Body into the correct pet type. This allows "polymorphism" in the
	// pet blocks.
	pets := []Pet{}
	for _, p := range petsHCL.PetHCLBodies {
		switch petType := p.Type; petType {
		case "cat":
			cat := &Cat{Name: p.Name, Sound: defaultCatSound}
			if p.CharacteristicsHCL != nil {
				if diag := gohcl.DecodeBody(p.CharacteristicsHCL.HCL, evalContext, cat); diag.HasErrors() {
					return []Pet{}, fmt.Errorf(
						"error in ReadConfig decoding cat HCL configuration: %w", diag,
					)
				}
			}
			pets = append(pets, cat)
		case "dog":
			dog := &Dog{Name: p.Name, Breed: defaultDogBreed}
			if p.CharacteristicsHCL != nil {
				if diag := gohcl.DecodeBody(p.CharacteristicsHCL.HCL, evalContext, dog); diag.HasErrors() {
					return []Pet{}, fmt.Errorf(
						"error in ReadConfig decoding dog HCL configuration: %w", diag,
					)
				}
			}
			pets = append(pets, dog)
		default:
			// Error in the case of an unknown type. In the future, more types
			// could be added to the switch to support, for example, fish
			// owners.
			return []Pet{}, fmt.Errorf("error in ReadConfig: unknown pet type `%s`", petType)
		}
	}
	return pets, nil
}

// createContext is a helper function that creates an *hcl.EvalContext to be
// used in decoding HCL. It creates a set of variables at env.KEY
// (namely, CAT_SOUND). It also creates a function "random(...string)" that can
// be used to assign a random value in an HCL config.
func createContext() (*hcl.EvalContext, error) {
	// Extract the sound cats make from the environment, with a default.
	catSound := defaultCatSound
	if os.Getenv(catSoundKey) != "" {
		catSound = os.Getenv(catSoundKey)
	}

	// variables is a list of cty.Value for use in Decoding HCL. These will
	// be nested by using ObjectVal as a value. For istance:
	//   env.CAT_SOUND => "meow"
	variables := map[string]cty.Value{
		environmentKey: cty.ObjectVal(map[string]cty.Value{
			catSoundKey: cty.StringVal(catSound),
		}),
	}

	// functions is a list of cty.Functions for use in Decoding HCL.
	functions := map[string]function.Function{
		"random": function.New(&function.Spec{
			// Params represents required positional arguments, of which random
			// has none.
			Params: []function.Parameter{
				function.Parameter{Type: cty.Number, Name: "count"},
				function.Parameter{Type: cty.String, Name: "seperator"},
			},
			// VarParam allows a "VarArgs" type input, in this case, of
			// strings.
			VarParam: &function.Parameter{Type: cty.String},
			// Type is used to determine the output type from the inputs. In
			// the case of Random it only accepts strings and only returns
			// strings.
			Type: function.StaticReturnType(cty.String),
			// Impl is the actual function. First a count, then a seperating
			// character, then A "VarArgs" number of cty.String will be passed
			// in and a random count returned, joined by seperator, will be
			// returned as a cty.String.
			Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
				// The first argument is a number, the count of random items to
				// select. Get that count as an int64.
				cCount := args[0]
				count, _ := cCount.AsBigFloat().Int64()

				// The second argument is a string, with which to join the
				// selected iteams. Get that seperator as a string.
				cSep := args[1]
				sep := cSep.AsString()

				// Now discard the first two arguments, giving us the VarArgs
				// list of strings to select from.
				args = args[2:]

				// Validate there are enough to satisfy the request.
				if int64(len(args)) < count {
					return cty.NilVal,
						fmt.Errorf(
							"unable to select %d random elements from list of length %d",
							count,
							len(args),
						)
				}

				// Grab a random element from the VarArgs. Add it to the
				// response and remove it, so it cannot be selected twice.
				resp := ""
				for i := 0; int64(i) < count; i++ {
					idx := rand.Intn(len(args))
					newString := args[idx].AsString()
					if resp == "" {
						resp = newString
					} else {
						resp = strings.Join([]string{resp, newString}, sep)
					}
					args = append(args[:idx], args[idx+1:]...)
				}

				// Return a StringVal. For instance:
				//   random(3, ", ", "a", "b", "c") => "c, a, b"
				return cty.StringVal(resp), nil
			},
		}),
	}

	// Return the constructed hcl.EvalContext.
	return &hcl.EvalContext{
		Variables: variables,
		Functions: functions,
	}, nil
}
