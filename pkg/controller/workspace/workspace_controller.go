package workspace

import (
	"context"

	work8spacev1alpha1 "github.com/infracloudio/work8spaces/pkg/apis/work8space/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_workspace")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Workspace Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWorkspace{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("workspace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Workspace
	err = c.Watch(&source.Kind{Type: &work8spacev1alpha1.Workspace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Namespaces and requeue the owner Workspace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &work8spacev1alpha1.Workspace{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWorkspace implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWorkspace{}

// ReconcileWorkspace reconciles a Workspace object
type ReconcileWorkspace struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Workspace object and makes changes based on the state read
// and what is in the Workspace.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Ns as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWorkspace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Workspace")

	// Fetch the Workspace instance
	instance := &work8spacev1alpha1.Workspace{}
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

	//Define a new Ns object
	ns := newNsForCR(instance)

	// Set Workspace instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, ns, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Ns already exists
	found := &corev1.Namespace{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: ns.Name, Namespace: ns.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Namespace:", "Namespace.Name", ns.Name)
		err = r.client.Create(context.TODO(), ns)
		if err != nil {
			return reconcile.Result{}, err
		}
		// Ns created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}
	// Ns already exists - don't requeue
	reqLogger.Info("Skip reconcile: Namespace already exists:", "Namespace.Name", found.Name)
	return reconcile.Result{}, nil
}

// newNsForCR returns a ns with the same name/namespace as the cr
func newNsForCR(cr *work8spacev1alpha1.Workspace) *corev1.Namespace {
	labels := map[string]string{
		"work8space.infracloud.io/workspace": cr.Name,
	}
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   cr.Name,
			Labels: labels,
		},
	}
}
