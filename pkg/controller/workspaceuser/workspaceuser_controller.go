package workspaceuser

import (
	"context"

	work8spacev1alpha1 "github.com/infracloudio/work8spaces/pkg/apis/work8space/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
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

var log = logf.Log.WithName("controller_workspaceuser")

const (
	// TODO(bhavin192): make this configurable
	saDefaultNamespace = "work8spaces-users"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new WorkspaceUser Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWorkspaceUser{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("workspaceuser-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource WorkspaceUser
	err = c.Watch(&source.Kind{Type: &work8spacev1alpha1.WorkspaceUser{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource ServiceAccount and
	// requeue the owner WorkspaceUser
	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &work8spacev1alpha1.WorkspaceUser{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resouce RoleBinding and
	// requeue the owner WorkspaceUser
	err = c.Watch(&source.Kind{Type: &rbacv1beta1.RoleBinding{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &work8spacev1alpha1.WorkspaceUser{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWorkspaceUser implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWorkspaceUser{}

// ReconcileWorkspaceUser reconciles a WorkspaceUser object
type ReconcileWorkspaceUser struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a WorkspaceUser object and makes changes based on the state read
// and what is in the WorkspaceUser.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWorkspaceUser) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling WorkspaceUser")

	// Fetch the WorkspaceUser instance
	user := &work8spacev1alpha1.WorkspaceUser{}
	err := r.client.Get(context.TODO(), request.NamespacedName, user)
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

	// Define a new ServiceAccount object
	sa := newServiceAccount(user)

	// Set WorkspaceUser instance as the owner and controller
	if err := controllerutil.SetControllerReference(user, sa, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this ServiceAccount already exists
	found := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sa.Name, Namespace: sa.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", sa.Namespace, "ServiceAccount.Name", sa.Name)
		err = r.client.Create(context.TODO(), sa)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// ServiceAccount already exists - don't requeue
	reqLogger.Info("Skip reconcile: ServiceAccount already exists", "ServiceAccount.Namespace", found.Namespace, "ServiceAccount.Name", found.Name)

	for _, wsItem := range user.Spec.Workspaces {
		ws := &work8spacev1alpha1.Workspace{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: wsItem.Name, Namespace: ""}, ws)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("Workspace not found", "Workspace.Name", wsItem.Name)
				// TODO(bhavin192): create an Event so
				// that this error is visible to users

				// Don't requeue until we get event
				// for Workspace creation
				// return reconcile.Result{}, nil

				// Not watching Workspace resouce
				// right now
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, err
		}

		// TODO(bhavin192): is passing sa object correct here?
		rb := createRoleBinding(sa, wsItem)

		// Set WorkspaceUser instance as the owner and controller
		if err = controllerutil.SetControllerReference(user, rb, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		rbFound := &rbacv1beta1.RoleBinding{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: rb.Name, Namespace: rb.ObjectMeta.Namespace}, rbFound)
		if err != nil && errors.IsNotFound(err) {
			reqLogger.Info("Creating a new RoleBinding", "RoleBinding.Namespace", rb.Namespace, "RoleBinding.Name", rb.Name)
			if err = r.client.Create(context.TODO(), rb); err != nil {
				return reconcile.Result{}, err
			}
		} else if err != nil {
			return reconcile.Result{}, err
		}

		// TODO(bhavin192): Update logic for RB
	}

	return reconcile.Result{}, nil
}

// newServiceAccount returns a serviceaccount created in
// saDefaultNamespace with same name as of WorkspaceUser
func newServiceAccount(user *work8spacev1alpha1.WorkspaceUser) *corev1.ServiceAccount {
	labels := map[string]string{
		"work8space.infracloud.io/user": user.Name,
	}
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      user.Name,
			Namespace: saDefaultNamespace,
			Labels:    labels,
		},
		// TODO(bhavin192): check if this is required or not
		// AutomountServiceAccountToken: true,
	}

}

func createRoleBinding(sa *corev1.ServiceAccount, wsItem work8spacev1alpha1.WUWorkspaceItem) *rbacv1beta1.RoleBinding {
	labels := map[string]string{
		"work8space.infracloud.io/user":      sa.Name,
		"work8space.infracloud.io/workspace": wsItem.Name,
	}
	rbName := sa.Name + "-" + wsItem.Name
	return &rbacv1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rbName,
			Namespace: wsItem.Name,
			Labels:    labels,
		},
		Subjects: []rbacv1beta1.Subject{{
			// TODO(bhavin192): is this approach correct
			// Kind: sa.Kind,
			Kind:      "ServiceAccount",
			Name:      sa.Name,
			Namespace: sa.Namespace,
		}},
		RoleRef: rbacv1beta1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     wsItem.Role,
		},
	}
}
