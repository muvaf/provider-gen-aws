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
package codegen

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
)

type ControllerGeneratorChain []func(*generate.Generator, string) error

func (a ControllerGeneratorChain) Generate(g *generate.Generator, controllerPath string) error {
	for _, f := range a {
		if err := f(g, controllerPath); err != nil {
			return err
		}
	}
	return nil
}

type ControllerGeneratorFn func(*generate.Generator, string) error

func (a ControllerGeneratorFn) Generate(g *generate.Generator, controllerPath string) error {
	return a(g, controllerPath)
}

func GenerateController(g *generate.Generator, controllerPath string) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		// TODO(muvaf): "manager" is hard-coded in ACK.
		b, err := g.GenerateCRDResourcePackageFile(crd.Names.Original, "manager")
		if err != nil {
			return err
		}
		path := filepath.Join(controllerPath, fmt.Sprintf("%s.go", crd.Names.Snake))
		if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}
	return nil
}

func GenerateConversions(g *generate.Generator, controllerPath string) error {
	crds, err := g.GetCRDs()
	if err != nil {
		return err
	}
	for _, crd := range crds {
		// TODO(muvaf): "sdk" is hard-coded in ACK.
		b, err := g.GenerateCRDResourcePackageFile(crd.Names.Original, "sdk")
		if err != nil {
			return err
		}
		path := filepath.Join(controllerPath, fmt.Sprintf("%s_conversions.go", crd.Names.Snake))
		if err := ioutil.WriteFile(path, b.Bytes(), 0666); err != nil {
			return err
		}
	}
	return nil
}
