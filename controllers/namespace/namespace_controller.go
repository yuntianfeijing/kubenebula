package namespace

import (
	"context"
	"encoding/base64"
	//appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubenebula.io/kubenebula/api/tenant/v1alpha1"
	"kubenebula.io/kubenebula/constants"

	"kubenebula.io/kubenebula/utils/sliceutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	adminAlias           = "命名空间管理员角色"
	developerAlias       = "命名空间操作员角色"
	viewerAlias          = "命名空间观察员角色"
	adminDescription     = "拥有命名空间的所有资源的管理权限"
	developerDescription = "拥有命名空间的除角色管理以外的所有资源的管理权限"
	viewerDescription    = "拥有命名空间所有资源的查看权限"
)

var (
	admin = rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "admin",
			Labels: map[string]string{constants.ResourceLabel: constants.ResourceRole},
			Annotations: map[string]string{
				constants.CreatorAnnotationKey:     constants.System,
				constants.DisplayNameAnnotationKey: adminAlias,
				constants.DescriptionAnnotationKey: adminDescription}},
		Rules: []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}}
	developer = rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "developer",
			Labels: map[string]string{constants.ResourceLabel: constants.ResourceRole},
			Annotations: map[string]string{
				constants.CreatorAnnotationKey:     constants.System,
				constants.DisplayNameAnnotationKey: developerAlias,
				constants.DescriptionAnnotationKey: developerDescription}},
		Rules: []rbac.PolicyRule{
			{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}},
			{Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions", "batch", "autoscaling", "app.k8s.io", "monitoring.coreos.com", "networking.k8s.io"}, Resources: []string{"*"}}}}
	viewer = rbac.Role{ObjectMeta: metav1.ObjectMeta{
		Name:   "viewer",
		Labels: map[string]string{constants.ResourceLabel: constants.ResourceRole},
		Annotations: map[string]string{
			constants.CreatorAnnotationKey:     constants.System,
			constants.DisplayNameAnnotationKey: viewerAlias,
			constants.DescriptionAnnotationKey: viewerDescription}},
		Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}}
	defaultRoles = []rbac.Role{admin, developer, viewer}
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Namespace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &NamespaceReconcile{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to Namespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &NamespaceReconcile{}

// NamespaceReconcile reconciles a Namespace object
type NamespaceReconcile struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Namespace object and makes changes based on the state read
// and what is in the Namespace.Spec
// +kubebuilder:rbac:groups=core.kubenebula.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubenebula.io,resources=namespaces/status,verbs=get;update;patch
func (r *NamespaceReconcile) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Namespace instance
	instance := &corev1.Namespace{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// The object is being deleted
			// our finalizer is present, so lets handle our external dependency
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// name of your custom finalizer
	finalizer := "finalizers.kubenebula.io/namespaces"

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			// TODO add some thing
			//if err = r.deleteRouter(instance.Name); err != nil {
			//	return reconcile.Result{}, err
			//}
			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}
	if err = r.chechOrAddTeamLabel(instance); err != nil {
		return reconcile.Result{}, err
	}
	controlledByTeam, err := r.isControlledByTeam(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !controlledByTeam {
		err = r.deleteRoleBindings(instance)
		return reconcile.Result{}, err
	}

	//if err = r.checkAndBindTeam(instance); err != nil {
	//	return reconcile.Result{}, err
	//}

	if err = r.checkAndCreateRoles(instance); err != nil {
		return reconcile.Result{}, err
	}

	//if err = r.checkAndCreateRoleBindings(instance); err != nil {
	//	return reconcile.Result{}, err
	//}
	return reconcile.Result{}, nil
}

func (r *NamespaceReconcile) isControlledByTeam(namespace *corev1.Namespace) (bool, error) {
	// without team label
	if teamName, ok := namespace.Labels[constants.TeamLabelKey]; !ok || (teamName == "") {
		return false, nil
	}
	return true, nil
}

// Create default roles
func (r *NamespaceReconcile) checkAndCreateRoles(namespace *corev1.Namespace) error {
	for _, role := range defaultRoles {
		found := &rbac.Role{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: role.Name}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				role := role.DeepCopy()
				role.Namespace = namespace.Name
				err = r.Create(context.TODO(), role)
				if err != nil {
					klog.Error(err)
					return err
				}
				return nil
			} else {
				klog.Error(err)
				return err
			}
		}
		if !reflect.DeepEqual(found.Rules, role.Rules) {
			found.Rules = role.Rules
			if err := r.Update(context.TODO(), found); err != nil {
				klog.Error(err)
				return err
			}
		}
	}
	return nil
}

//只绑定创建者到admin
func (r *NamespaceReconcile) checkAndCreateRoleBindings(namespace *corev1.Namespace) error {
	return nil
}

/*
func (r *NamespaceReconcile) checkAndCreateRoleBindings(namespace *corev1.Namespace) error {

	teamName := namespace.Labels[constants.TeamLabelKey]
	creatorName := namespace.Annotations[constants.CreatorAnnotationKey]

	creator := rbac.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: creatorName}

	teamAdminBinding := &rbac.ClusterRoleBinding{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("team:%s:admin", teamName)}, teamAdminBinding)

	if err != nil {
		return err
	}

	adminBinding := &rbac.RoleBinding{}
	adminBinding.Name = admin.Name
	adminBinding.Namespace = namespace.Name
	adminBinding.RoleRef = rbac.RoleRef{Name: admin.Name, APIGroup: "rbac.authorization.k8s.io", Kind: "Role"}
	adminBinding.Subjects = teamAdminBinding.Subjects

	if creator.Name != "" {
		if adminBinding.Subjects == nil {
			adminBinding.Subjects = make([]rbac.Subject, 0)
		}
		if !k8sutil.ContainsUser(adminBinding.Subjects, creatorName) {
			adminBinding.Subjects = append(adminBinding.Subjects, creator)
		}
	}

	found := &rbac.RoleBinding{}

	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: adminBinding.Name}, found)

	if errors.IsNotFound(err) {
		err = r.Create(context.TODO(), adminBinding)
		if err != nil {
			klog.Errorf("creating role binding namespace: %s,role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
			return err
		}
		found = adminBinding
	} else if err != nil {
		klog.Errorf("get role binding namespace: %s,role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
		return err
	}

	if !reflect.DeepEqual(found.RoleRef, adminBinding.RoleRef) {
		err = r.Delete(context.TODO(), found)
		if err != nil {
			klog.Errorf("deleting role binding namespace: %s, role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
			return err
		}
		err = fmt.Errorf("conflict role binding %s.%s, waiting for recreate", namespace.Name, adminBinding.Name)
		klog.Errorf("conflict role binding namespace: %s, role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
		return err
	}

	if !reflect.DeepEqual(found.Subjects, adminBinding.Subjects) {
		found.Subjects = adminBinding.Subjects
		err = r.Update(context.TODO(), found)
		if err != nil {
			klog.Errorf("updating role binding namespace: %s, role binding: %s, error: %s", namespace.Name, adminBinding.Name, err)
			return err
		}
	}

	teamViewerBinding := &rbac.ClusterRoleBinding{}

	err = r.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("team:%s:viewer", teamName)}, teamViewerBinding)

	if err != nil {
		return err
	}

	viewerBinding := &rbac.RoleBinding{}
	viewerBinding.Name = viewer.Name
	viewerBinding.Namespace = namespace.Name
	viewerBinding.RoleRef = rbac.RoleRef{Name: viewer.Name, APIGroup: "rbac.authorization.k8s.io", Kind: "Role"}
	viewerBinding.Subjects = teamViewerBinding.Subjects

	err = r.Get(context.TODO(), types.NamespacedName{Namespace: namespace.Name, Name: viewerBinding.Name}, found)

	if errors.IsNotFound(err) {
		err = r.Create(context.TODO(), viewerBinding)
		if err != nil {
			klog.Errorf("creating role binding namespace: %s, role binding: %s, error: %s", namespace.Name, viewerBinding.Name, err)
			return err
		}
		found = viewerBinding
	} else if err != nil {
		return err
	}

	if !reflect.DeepEqual(found.RoleRef, viewerBinding.RoleRef) {
		err = r.Delete(context.TODO(), found)
		if err != nil {
			klog.Errorf("deleting conflict role binding namespace: %s, role binding: %s, %s", namespace.Name, viewerBinding.Name, err)
			return err
		}
		err = fmt.Errorf("conflict role binding %s.%s, waiting for recreate", namespace.Name, viewerBinding.Name)
		klog.Errorf("conflict role binding namespace: %s, role binding: %s, error: %s", namespace.Name, viewerBinding.Name, err)
		return err
	}

	if !reflect.DeepEqual(found.Subjects, viewerBinding.Subjects) {
		found.Subjects = viewerBinding.Subjects
		err = r.Update(context.TODO(), found)
		if err != nil {
			klog.Errorf("updating role binding namespace: %s, role binding: %s, error: %s", namespace.Name, viewerBinding.Name, err)
			return err
		}
	}

	return nil
}
*/
func (r *NamespaceReconcile) checkAndBindTeam(namespace *corev1.Namespace) error {

	teamName := namespace.Labels[constants.TeamLabelKey]

	if teamName == "" {
		return nil
	}

	team := &v1alpha1.Team{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: teamName}, team)

	if err != nil {
		// skip if team not found
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("bind team namespace: %s, team: %s, error: %s", namespace.Name, teamName, err)
		return err
	}

	if !metav1.IsControlledBy(namespace, team) {
		if err := controllerutil.SetControllerReference(team, namespace, r.Scheme); err != nil {
			klog.Errorf("bind team namespace: %s, team: %s, error: %s", namespace.Name, teamName, err)
			return err
		}
		err = r.Update(context.TODO(), namespace)
		if err != nil {
			klog.Errorf("bind team namespace: %s, team: %s, error: %s", namespace.Name, teamName, err)
			return err
		}
	}

	return nil
}

func (r *NamespaceReconcile) deleteRoleBindings(namespace *corev1.Namespace) error {
	klog.V(4).Info("deleting role bindings namespace: ", namespace.Name)
	adminBinding := &rbac.RoleBinding{}
	adminBinding.Name = admin.Name
	adminBinding.Namespace = namespace.Name
	err := r.Delete(context.TODO(), adminBinding)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("deleting role binding namespace: %s, role binding: %s,error: %s", namespace.Name, adminBinding.Name, err)
		return err
	}
	developerBinding := &rbac.RoleBinding{}
	developerBinding.Name = developer.Name
	developerBinding.Namespace = namespace.Name
	err = r.Delete(context.TODO(), developerBinding)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("deleting role binding namespace: %s, role binding: %s,error: %s", namespace.Name, developerBinding.Name, err)
		return err
	}
	viewerBinding := &rbac.RoleBinding{}
	viewerBinding.Name = viewer.Name
	viewerBinding.Namespace = namespace.Name
	err = r.Delete(context.TODO(), viewerBinding)
	if err != nil && !errors.IsNotFound(err) {
		klog.Errorf("deleting role binding namespace: %s,role binding: %s,error: %s", namespace.Name, viewerBinding.Name, err)
		return err
	}
	return nil
}
func (r *NamespaceReconcile) chechOrAddTeamLabel(namespace *corev1.Namespace) (err error) {
	team := ""
	if namespace.Annotations != nil {
		if value, ok := namespace.Annotations[constants.TeamAnnotationKey]; ok {
			team = value
		}
	}
	if team != "" {
		if namespace.Labels == nil {
			namespace.Labels = make(map[string]string)
			namespace.Labels[constants.TeamLabelKey] = base64.RawURLEncoding.EncodeToString([]byte(team))
			if err := r.Update(context.Background(), namespace); err != nil {
				return err
			}
		}
		if value, ok := namespace.Labels[constants.TeamLabelKey]; !ok || (value == "") {
			namespace.Labels[constants.TeamLabelKey] = base64.RawURLEncoding.EncodeToString([]byte(team))
			if err := r.Update(context.Background(), namespace); err != nil {
				return err
			}
		}
	}

	return nil
}

/*
func (r *NamespaceReconcile) deleteRouter(namespace string) error {
	routerName := constants.IngressControllerPrefix + namespace

	// delete service first
	found := corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
	}

	err = r.Delete(context.TODO(), &found)
	if err != nil {
		klog.Error(err)
		return err
	}

	// delete deployment
	deploy := appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	err = r.Delete(context.TODO(), &deploy)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil
}
*/
