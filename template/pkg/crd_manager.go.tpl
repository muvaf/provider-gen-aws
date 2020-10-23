{{ template "boilerplate" }}

package {{ .CRD.Names.Snake }}

import (
	"context"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	svcsdkapi "github.com/aws/aws-sdk-go/service/{{ .ServiceIDClean }}/{{ .ServiceIDClean }}iface"
	svcapi "github.com/aws/aws-sdk-go/service/{{ .ServiceIDClean }}"
	"github.com/aws/aws-sdk-go/aws/session"
	ackerr "github.com/aws/aws-controllers-k8s/pkg/errors"

	"github.com/crossplane/provider-template/apis/{{ .ServiceIDClean }}/{{ .APIVersion}}"
	awsclient "github.com/crossplane/provider-template/pkg/client"
)

const (
	errUnexpectedObject = "managed resource is not an repository resource"

	errCreateSession = "cannot create a new session"
	errCreate = "cannot create {{ .CRD.Names.Camel }} in AWS"
)

// Setup{{ .CRD.Names.Camel }} adds a controller that reconciles {{ .CRD.Names.Camel }}.
func Setup{{ .CRD.Names.Camel }}(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName({{ .APIVersion}}.{{ .CRD.Names.Camel }}GroupKind)
	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&{{ .APIVersion}}.{{ .CRD.Names.Camel }}{}).
		Complete(managed.NewReconciler(mgr,
			cpresource.ManagedKind({{ .APIVersion}}.{{ .CRD.Names.Camel }}GroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithConnectionPublishers(),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube client.Client
}

func (c *connector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*{{ .APIVersion}}.{{ .CRD.Names.Camel }})
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := awsclient.GetConfig(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, err
	}
	sess, err := session.NewSession(cfg)
  return &external{client: svcapi.New(sess), kube: c.kube}, errors.Wrap(err, errCreateSession)
}

type external struct {
	kube   client.Client
	client svcsdkapi.{{ .SDKAPIInterfaceTypeName }}API
}

func (e *external) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*{{ .APIVersion}}.{{ .CRD.Names.Camel }})
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	cr.SetConditions(runtimev1alpha1.Available())
  {{- if .CRD.Ops.ReadMany }}
	input := Generate{{ .CRD.Ops.ReadMany.InputRef.Shape.ShapeName }}(cr)
	// TODO(muvaf): Generated conversion code has logic about the input, like if(len...)
	// and has a return statement we can't control here in an arbitrary function.
	// TODO(muvaf): Generated code has an assumption about the module name of the type (svcapitypes)
	// but that doesn't always hold true.
{{ $setCode := GoCodeSetReadManyOutput .CRD "resp" "cr" 1 false }}
	{{ if not ( Empty $setCode ) }}resp{{ else }}_{{ end }}, err := e.client.{{ .CRD.Ops.ReadMany.Name }}WithContext(ctx, input)
	if err != nil {
		if awsErr, ok := ackerr.AWSError(err); ok && awsErr.Code() == "{{ ResourceExceptionCode .CRD 404 }}" {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, err
	}
{{ $setCode }}
{{- end }}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg cpresource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*{{ .APIVersion}}.{{ .CRD.Names.Camel }})
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(runtimev1alpha1.Creating())
	input := Generate{{ .CRD.Ops.Create.InputRef.Shape.ShapeName }}(cr)
  {{ $createCode := GoCodeSetCreateOutput .CRD "resp" "cr" 1 false }}
  	{{ if and .CRD.StatusFields ( not ( Empty $createCode ) ) }}resp{{ else }}_{{ end }}, err := e.client.{{ .CRD.Ops.Create.Name }}WithContext(ctx, input)
  	if err != nil {
  	  return managed.ExternalCreation{}, errors.Wrap(err, errCreate)
  	}
  {{ $createCode }}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg cpresource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*{{ .APIVersion}}.{{ .CRD.Names.Camel }})
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	cr.SetConditions(runtimev1alpha1.Available())
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg cpresource.Managed) error {
	cr, ok := mg.(*{{ .APIVersion}}.{{ .CRD.Names.Camel }})
	if !ok {
		return errors.New(errUnexpectedObject)
	}
	cr.Status.SetConditions(runtimev1alpha1.Deleting())

	{{- if .CRD.Ops.Delete }}
  	input := Generate{{ .CRD.Ops.Delete.InputRef.Shape.ShapeName }}(cr)
  	_, err := e.client.{{ .CRD.Ops.Delete.Name }}WithContext(ctx, input)
  	// TODO(muvaf): Ignore NotFound errors here.
  	return errors.Wrap(err, errCreate)
  {{- else }}
  	// TODO(jaypipes): Figure this out...
  	return nil
  {{ end }}
}
