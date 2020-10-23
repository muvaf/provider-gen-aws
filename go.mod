module github.com/crossplane/provider-template

go 1.13

require (
	github.com/aws/aws-controllers-k8s v0.0.0-20201022191406-64428498d932
	github.com/crossplane/crossplane-runtime v0.10.0
	github.com/crossplane/crossplane-tools v0.0.0-20201022234345-cea4faeaa0bb
	github.com/google/go-cmp v0.4.0
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/pkg/errors v0.9.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.18.6
	k8s.io/apimachinery v0.18.6
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.4.0
)
