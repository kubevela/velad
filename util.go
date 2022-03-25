package main

import (
	"io"
	"io/fs"
	"io/ioutil"
	"path"
	"strings"

	"github.com/oam-dev/kubevela/references/cli"
)

// SaveToTemp helps save an embedded file into a temporary file
func SaveToTemp(file fs.File, format string) (string, error) {
	tempFile, err := ioutil.TempFile("/var", format)
	if err != nil {
		return "", err
	}
	defer CloseQuietly(tempFile)

	_, err = io.Copy(tempFile, file)
	if err != nil {
		return "", err
	}
	return path.Join("/var", tempFile.Name()), nil
}

// CloseQuietly closes `io.Closer` quietly. Very handy and helpful for code
// quality too.
func CloseQuietly(d io.Closer) {
	_ = d.Close()
}

func TransArgsToString(args cli.InstallArgs) []string {
	var res []string
	if args.Values != nil {
		res = append(res, "--set="+strings.Join(args.Values, ","))
	}
	if args.Namespace != "" {
		res = append(res, "--namespace="+args.Namespace)
	}
	if !args.Detail {
		res = append(res, "--detail=false")
	}
	if !args.ReuseValues {
		res = append(res, "--reuse=false")
	}
	return res
}
