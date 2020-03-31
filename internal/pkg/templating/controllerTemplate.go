package templating

import (
	"bytes"
	"text/template"
)

// ControllerTemplateConfig contains the necessary fields to render the controller template
type ControllerTemplateConfig struct {
	Kind                 string
	LowerKind            string
	OwnerAPIVersion      string
	ImportMap            map[string]string
	ResourceWatches      []string
	ResourceForReconcile []string
}

// NewControllerTemplateConfig returns the necessary templating configuration
func NewControllerTemplateConfig(kind, lowerKind, ownerAPIVersion string, importMap map[string]string, resourceWatches, resourceForReconcile []string) *ControllerTemplateConfig {
	return &ControllerTemplateConfig{
		kind,
		lowerKind,
		ownerAPIVersion,
		importMap,
		resourceWatches,
		resourceForReconcile,
	}
}

// Execute renders the template and returns the templated string
func (c *ControllerTemplateConfig) Execute() (string, error) {
	temp, err := template.New("resourceFuncTemplate").Parse(c.GetTemplate())
	if err != nil {
		return "", err
	}

	var wr bytes.Buffer
	err = temp.Execute(&wr, c)
	if err != nil {
		return "", err
	}
	return wr.String(), nil
}

// GetTemplate returns the necessary template
func (c *ControllerTemplateConfig) GetTemplate() string {
	return `
		package {{ .LowerKind }}
		import (
			"context"
			{{range $p, $i := .ImportMap -}}
			{{$i}} "{{$p}}"
			{{end}}
		)
		var log = logf.Log.WithName("controller_{{ .LowerKind }}")
		/**
		* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
		* business logic.  Delete these comments after modifying this file.*
		*/
		// Add creates a new {{ .Kind }} Controller and adds it to the Manager. The Manager will set fields on the Controller
		// and Start it when the Manager is Started.
		func Add(mgr manager.Manager) error {
			return add(mgr, newReconciler(mgr))
		}
		// newReconciler returns a new reconcile.Reconciler
		func newReconciler(mgr manager.Manager) reconcile.Reconciler {
			return &Reconcile{{ .Kind }}{client: mgr.GetClient(), scheme: mgr.GetScheme()}
		}
		// add adds a new Controller to mgr with r as the reconcile.Reconciler
		func add(mgr manager.Manager, r reconcile.Reconciler) error {
			// Create a new controller
			c, err := controller.New("{{ .LowerKind }}-controller", mgr, controller.Options{Reconciler: r})
			if err != nil {
				return err
			}
			// Watch for changes to primary resource {{ .Kind }}
			err = c.Watch(&source.Kind{Type: &{{ .OwnerAPIVersion }}.{{ .Kind }}{}}, &handler.EnqueueRequestForObject{})
			if err != nil {
				return err
			}
			// GENERATED BY CONVERSION KIT
		
			{{range $f :=  .ResourceWatches -}}
				{{$f}}
				
			{{end}}
		
			return nil
		}
		// blank assignment to verify that Reconcile{{ .Kind }} implements reconcile.Reconciler
		var _ reconcile.Reconciler = &Reconcile{{ .Kind }}{}
		// Reconcile{{ .Kind }} reconciles a {{ .Kind }} object
		type Reconcile{{ .Kind }} struct {
			// This client, initialized using mgr.Client() above, is a split client
			// that reads objects from the cache and writes to the apiserver
			client client.Client
			scheme *runtime.Scheme
		}
		// Reconcile reads that state of the cluster for a {{ .Kind }} object and makes changes based on the state read
		// and what is in the {{ .Kind }}.Spec
		// TODO(user): Modify this Reconcile function to implement your Controller logic.
		// Note:
		// The Controller will requeue the Request to be processed again if the returned error is non-nil or
		// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
		func (r *Reconcile{{ .Kind }}) Reconcile(request reconcile.Request) (reconcile.Result, error) {
			reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
			reqLogger.Info("Reconciling {{ .Kind }}")
			// Fetch the {{ .Kind }} instance
			instance := &{{ .OwnerAPIVersion }}.{{ .Kind }}{}
			err := r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				if errors.IsNotFound(err) {
					// Request object not found, could have been deleted after reconcile request.
					// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
					// Return and don't requeue
					return reconcile.Result{}, nil
				}
				// Error reading the object - requeue the request.
				return reconcile.Result{}, err
			}
			{{range $r := .ResourceForReconcile -}}
				{{$r}}
			{{end}}
		
			return reconcile.Result{}, nil
		}
	`
}