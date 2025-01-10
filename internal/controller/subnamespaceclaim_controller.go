package controller

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	subnsv1alpha1 "github.com/kraudcloud/subns-controller/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// SubNamespaceClaimReconciler reconciles a SubNamespaceClaim object
type SubNamespaceClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=subns.kraud.cloud,resources=subnamespaceclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=subns.kraud.cloud,resources=subnamespaceclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=subns.kraud.cloud,resources=subnamespaceclaims/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *SubNamespaceClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the SubNamespaceClaim instance
	subNamespaceClaim := &subnsv1alpha1.SubNamespaceClaim{}
	err := r.Get(ctx, req.NamespacedName, subNamespaceClaim)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Construct the full namespace name by prefixing with parent namespace
	fullNamespaceName := fmt.Sprintf("%s-%s", subNamespaceClaim.Namespace, subNamespaceClaim.Spec.Name)

	// Create or ensure namespace exists
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: fullNamespaceName,
			Labels: map[string]string{
				"parent-namespace": subNamespaceClaim.Namespace,
				"managed-by":       "subns-controller",
			},
		},
	}

	err = r.Get(ctx, client.ObjectKey{Name: namespace.Name}, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			if err = r.Create(ctx, namespace); err != nil {
				log.Error(err, "Failed to create namespace")
				return ctrl.Result{}, err
			}
			log.Info("Created namespace", "namespace", namespace.Name)
		} else {
			return ctrl.Result{}, err
		}
	}

	// Create or update each RoleBinding from the spec
	for _, rbTmpl := range subNamespaceClaim.Spec.RoleBindings {
		roleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      rbTmpl.Name,
				Namespace: fullNamespaceName,
			},
			Subjects: rbTmpl.Subjects,
			RoleRef:  rbTmpl.RoleRef,
		}

		err = r.Get(ctx, client.ObjectKey{Name: roleBinding.Name, Namespace: roleBinding.Namespace}, &rbacv1.RoleBinding{})
		if err != nil {
			if errors.IsNotFound(err) {
				if err = r.Create(ctx, roleBinding); err != nil {
					log.Error(err, "Failed to create RoleBinding")
					return ctrl.Result{}, err
				}
				log.Info("Created RoleBinding", "roleBinding", roleBinding.Name)
			} else {
				return ctrl.Result{}, err
			}
		} else {
			// RoleBinding exists, update it
			if err = r.Update(ctx, roleBinding); err != nil {
				log.Error(err, "Failed to update RoleBinding")
				return ctrl.Result{}, err
			}
			log.Info("Updated RoleBinding", "roleBinding", roleBinding.Name)
		}
	}

	// Update status
	subNamespaceClaim.Status.FullNamespace = fullNamespaceName
	condition := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		Reason:             "ResourcesCreated",
		Message:            "Namespace and RoleBindings created successfully",
	}

	subNamespaceClaim.Status.Conditions = []metav1.Condition{condition}
	if err := r.Status().Update(ctx, subNamespaceClaim); err != nil {
		log.Error(err, "Failed to update SubNamespaceClaim status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SubNamespaceClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&subnsv1alpha1.SubNamespaceClaim{}).
		Complete(r)
}
