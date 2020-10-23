/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/crossplane/provider-template/pkg/codegen"
)

const (
	awsSDKGithubURL = "https://raw.githubusercontent.com/aws/aws-sdk-go"
)

// Example call:
// go run -tags codegen cmd/generator/main.go --service ecr/2015-09-21 --version v1alpha1
// make generate

func main() {
	var (
		app     = kingpin.New(filepath.Base(os.Args[0]), "AWS Provider Generator.").DefaultEnvars()
		service = app.Flag("service", "The AWS service you want to generate controller for. "+
			"The format should be ecr/2015-09-21. "+
			"You can take a look at available services in https://github.com/aws/aws-sdk-go/tree/master/models/apis").Short('s').Required().String()
		sdkRevision = app.Flag("sdk-revision", "The revision hash of the AWS SDK Go to use to pull models.").Default("v1.34.32").String()
		// TODO(muvaf): ACK supports generation of all services in one group. But we need
		// to have CRD-level granularity so that we can version them separately.
		version = app.Flag("version", "The version for the generated API types.").Short('v').Required().String()

		providerDir = app.Flag("provider-dir", "The directory of the AWS Provider.").Short('o').Default(".").String()
		templateDir = app.Flag("template-dir", "The directory of the template to use.").Short('t').Default("template").String()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))
	sdkDir, err := prepareSDKFiles(*sdkRevision, *service)
	if err != nil {
		kingpin.Fatalf("cannot prepare sdk files: %s", err.Error())
	}
	s := strings.Split(*service, "/")
	if len(s) < 2 {
		kingpin.Fatalf("service argument does not conform the format <service name>/<service version date>")
	}
	g := codegen.NewGeneration(s[0], s[1], *version, *providerDir, sdkDir, *templateDir)
	kingpin.FatalIfError(g.Generate(), "api could not be generated")
}

func prepareSDKFiles(sdkRevision, service string) (string, error) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "cp-aws")
	if err != nil {
		return "", errors.Wrap(err, "cannot create a temp directory")
	}
	jsonDir := filepath.Join(tempDir, "models", "apis", service)
	if err := os.MkdirAll(jsonDir, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "cannot create folders in temp directory")
	}
	for _, filename := range []string{"api-2.json", "doc-2.json"} {
		url := fmt.Sprintf("%s/%s/models/apis/%s/%s", awsSDKGithubURL, sdkRevision, service, filename)
		resp, err := http.Get(url)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("cannot download %s", url))
		}

		file, err := os.Create(filepath.Join(jsonDir, filename))
		if err != nil {
			return "", errors.Wrap(err, "cannot create temp file")
		}
		if _, err := io.Copy(file, resp.Body); err != nil {
			return "", errors.Wrap(err, "cannot copy the body of response to temp file")
		}
		_ = resp.Body.Close()
		_ = file.Close()
	}
	return tempDir, nil
}
