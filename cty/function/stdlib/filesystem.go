package stdlib

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

// MakeFileExistsFunc is a function that takes a path
// and determines whether a file exists at that path
func MakeFileExistsFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Bool),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			path, err := homedir.Expand(path)
			if err != nil {
				return cty.UnknownVal(cty.Bool), fmt.Errorf("failed to expand ~: %s", err)
			}

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			// Ensure that the path is canonical for the host OS
			path = filepath.Clean(path)

			fi, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					return cty.False, nil
				}
				return cty.UnknownVal(cty.Bool), fmt.Errorf("failed to stat %s", path)
			}

			if fi.Mode().IsRegular() {
				return cty.True, nil
			}

			return cty.False, fmt.Errorf("%s is not a regular file, but %q",
				path, fi.Mode().String())
		},
	})
}

// MakeFileSetFunc is a function that takes a glob pattern
// and enumerates a file set from that pattern
func MakeFileSetFunc(baseDir string) function.Function {
	return function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
			{
				Name: "pattern",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Set(cty.String)),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			path := args[0].AsString()
			pattern := args[1].AsString()

			if !filepath.IsAbs(path) {
				path = filepath.Join(baseDir, path)
			}

			// Join the path to the glob pattern, while ensuring the full
			// pattern is canonical for the host OS. The joined path is
			// automatically cleaned during this operation.
			pattern = filepath.Join(path, pattern)

			matches, err := doublestar.Glob(pattern)
			if err != nil {
				return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to glob pattern (%s): %s", pattern, err)
			}

			var matchVals []cty.Value
			for _, match := range matches {
				fi, err := os.Stat(match)

				if err != nil {
					return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to stat (%s): %s", match, err)
				}

				if !fi.Mode().IsRegular() {
					continue
				}

				// Remove the path and file separator from matches.
				match, err = filepath.Rel(path, match)

				if err != nil {
					return cty.UnknownVal(cty.Set(cty.String)), fmt.Errorf("failed to trim path of match (%s): %s", match, err)
				}

				// Replace any remaining file separators with forward slash (/)
				// separators for cross-system compatibility.
				match = filepath.ToSlash(match)

				matchVals = append(matchVals, cty.StringVal(match))
			}

			if len(matchVals) == 0 {
				return cty.SetValEmpty(cty.String), nil
			}

			return cty.SetVal(matchVals), nil
		},
	})
}

// BasenameFunc is a function that takes a string containing a filesystem path
// and removes all except the last portion from it.
var BasenameFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(filepath.Base(args[0].AsString())), nil
	},
})

// DirnameFunc is a function that takes a string containing a filesystem path
// and removes the last portion from it.
var DirnameFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.StringVal(filepath.Dir(args[0].AsString())), nil
	},
})

// AbsPathFunc is a function that converts a filesystem path to an absolute path
var AbsPathFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		absPath, err := filepath.Abs(args[0].AsString())
		return cty.StringVal(filepath.ToSlash(absPath)), err
	},
})

// PathExpandFunc is a function that expands a leading ~ character to the current user's home directory.
var PathExpandFunc = function.New(&function.Spec{
	Params: []function.Parameter{
		{
			Name: "path",
			Type: cty.String,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {

		homePath, err := homedir.Expand(args[0].AsString())
		return cty.StringVal(homePath), err
	},
})

func readFileBytes(baseDir, path string) ([]byte, error) {
	path, err := homedir.Expand(path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand ~: %s", err)
	}

	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	// Ensure that the path is canonical for the host OS
	path = filepath.Clean(path)

	src, err := ioutil.ReadFile(path)
	if err != nil {
		// ReadFile does not return Terraform-user-friendly error
		// messages, so we'll provide our own.
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no file exists at %s", path)
		}
		return nil, fmt.Errorf("failed to read %s", path)
	}

	return src, nil
}

// FileExists determines whether a file exists at the given path.
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileExists(baseDir string, path cty.Value) (cty.Value, error) {
	fn := MakeFileExistsFunc(baseDir)
	return fn.Call([]cty.Value{path})
}

// FileSet enumerates a set of files given a glob pattern
//
// The underlying function implementation works relative to a particular base
// directory, so this wrapper takes a base directory string and uses it to
// construct the underlying function before calling it.
func FileSet(baseDir string, path, pattern cty.Value) (cty.Value, error) {
	fn := MakeFileSetFunc(baseDir)
	return fn.Call([]cty.Value{path, pattern})
}

// Basename takes a string containing a filesystem path and removes all except the last portion from it.
//
// The underlying function implementation works only with the path string and does not access the filesystem itself.
// It is therefore unable to take into account filesystem features such as symlinks.
//
// If the path is empty then the result is ".", representing the current working directory.
func Basename(path cty.Value) (cty.Value, error) {
	return BasenameFunc.Call([]cty.Value{path})
}

// Dirname takes a string containing a filesystem path and removes the last portion from it.
//
// The underlying function implementation works only with the path string and does not access the filesystem itself.
// It is therefore unable to take into account filesystem features such as symlinks.
//
// If the path is empty then the result is ".", representing the current working directory.
func Dirname(path cty.Value) (cty.Value, error) {
	return DirnameFunc.Call([]cty.Value{path})
}

// Pathexpand takes a string that might begin with a `~` segment, and if so it replaces that segment with
// the current user's home directory path.
//
// The underlying function implementation works only with the path string and does not access the filesystem itself.
// It is therefore unable to take into account filesystem features such as symlinks.
//
// If the leading segment in the path is not `~` then the given path is returned unmodified.
func Pathexpand(path cty.Value) (cty.Value, error) {
	return PathExpandFunc.Call([]cty.Value{path})
}
