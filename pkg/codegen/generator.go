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
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-controllers-k8s/pkg/generate"
	"github.com/pkg/errors"

	"github.com/aws/aws-controllers-k8s/pkg/model"
)

func WithGeneratorConfigFilePath(path string) GenerationOption {
	return func(g *Generation) {
		g.GeneratorConfigFilePath = path
	}
}

type GenerationOption func(*Generation)

type APIFileGenerator interface {
	Generate(g *generate.Generator, apiPath string) error
}

type ControllerGenerator interface {
	Generate(g *generate.Generator, controllerPath string) error
}

func NewGeneration(serviceName, serviceVersion, apiVersion, providerDirectory, sdkBasePath, templatePath string, opts ...GenerationOption) *Generation {
	g := &Generation{
		ServiceName:       serviceName,
		ServiceVersion:    serviceVersion,
		APIVersion:        apiVersion,
		ProviderDirectory: providerDirectory,
		SDKBasePath:       sdkBasePath,
		TemplateBasePath:  templatePath,
		apis: APIFileGeneratorChain{
			GenerateCRDFiles,
			GenerateTypesFile,
			GenerateEnumsFile,
			GenerateGroupVersionInfoFile,
			GenerateDocFile,
		},
		controller: ControllerGeneratorChain{
			GenerateController,
			GenerateConversions,
		},
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

type Generation struct {
	ServiceName             string
	ServiceVersion          string
	APIVersion              string
	ProviderDirectory       string
	SDKBasePath             string
	TemplateBasePath        string
	GeneratorConfigFilePath string

	apis       APIFileGenerator
	controller ControllerGenerator
}

func (g *Generation) Generate() error {
	apiPath := filepath.Join(g.ProviderDirectory, "apis", strings.Split(g.ServiceName, "/")[0], g.APIVersion)
	controllerPath := filepath.Join(g.ProviderDirectory, "pkg", "controller", strings.Split(g.ServiceName, "/")[0])
	sdkHelper := model.NewSDKHelper(g.SDKBasePath)
	sdkAPI, err := sdkHelper.API(g.ServiceName)
	if err != nil {
		return errors.Wrap(err, "cannot get the API model for service")
	}
	o, err := generate.New(sdkAPI, g.APIVersion, g.GeneratorConfigFilePath, g.TemplateBasePath)
	if err != nil {
		return errors.Wrap(err, "cannot create a new ACK Generator")
	}

	for _, path := range []string{apiPath, controllerPath} {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return errors.Wrap(err, "cannot create api folder")
		}
	}
	// TODO(muvaf): ACK generator requires all template files to be present during
	// initTemplates even though we don't use them.
	if err := g.apis.Generate(o, apiPath); err != nil {
		return errors.Wrap(err, "cannot generate API files")
	}

	if err := g.controller.Generate(o, controllerPath); err != nil {
		return errors.Wrap(err, "cannot generate controller files")
	}
	return nil
}
