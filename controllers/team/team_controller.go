/*
Copyright 2019 The KubeNebula authors.

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

package team

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"kubenebula.io/kubenebula/constants"
	"kubenebula.io/kubenebula/utils/sliceutil"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tenantv1alpha1 "kubenebula.io/kubenebula/api/tenant/v1alpha1"
)

const (
	teamAdminDescription   = "Allows admin access to perform any action on any resource, it gives full control over every resource in the team."
	teamRegularDescription = "Normal user in the team, can create namespace and DevOps project."
	teamViewerDescription  = "Allows viewer access to view all resources in the team."
)

var log = logf.Log.WithName("team-controller")

// TeamReconciler reconciles a Team object
type TeamReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tenant.kubenebula.io,resources=teams,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tenant.kubenebula.io,resources=teams/status,verbs=get;update;patch

func (r *TeamReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	//_ = context.Background()
	//_ = r.Log.WithValues("team", req.NamespacedName)
	instance := &tenantv1alpha1.Team{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// name of your custom finalizer
	finalizer := "finalizers.tenant.kubenebula.io"
	if instance.ObjectMeta.DeletionTimestamp.IsZero() { //被创建
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizer)
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else { //被删除
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			// our finalizer is present, so lets handle our external dependency
			//删除依赖的资源
			//if err := r.deleteDevOpsProjects(instance); err != nil {
			//	return reconcile.Result{}, err
			//}
			log.Info("Delete team todo some thing", "team", instance.Name, "name", instance.Name)
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
	//创建相关资源
	if err = r.createTeamAdmin(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createTeamRegular(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createTeamViewer(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.createTeamRoleBindings(instance); err != nil {
		return reconcile.Result{}, err
	}

	if err = r.bindNamespaces(instance); err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *TeamReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1alpha1.Team{}).
		Complete(r)
}

func (r *TeamReconciler) createTeamAdmin(instance *tenantv1alpha1.Team) error {
	found := &rbac.ClusterRole{}

	admin := getTeamAdmin(instance.Name)

	if err := controllerutil.SetControllerReference(instance, admin, r.Scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: admin.Name}, found)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating team role", "team", instance.Name, "name", admin.Name)
		err = r.Create(context.TODO(), admin)
		if err != nil {
			return err
		}
		found = admin
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(admin.Rules, found.Rules) || !reflect.DeepEqual(admin.Labels, found.Labels) || !reflect.DeepEqual(admin.Annotations, found.Annotations) {
		found.Rules = admin.Rules
		found.Labels = admin.Labels
		found.Annotations = admin.Annotations
		log.Info("Updating team role", "team", instance.Name, "name", admin.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TeamReconciler) createTeamRegular(instance *tenantv1alpha1.Team) error {
	found := &rbac.ClusterRole{}

	regular := getTeamRegular(instance.Name)

	if err := controllerutil.SetControllerReference(instance, regular, r.Scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: regular.Name}, found)

	if err != nil && errors.IsNotFound(err) {

		log.Info("Creating team role", "team", instance.Name, "name", regular.Name)
		err = r.Create(context.TODO(), regular)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		found = regular
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(regular.Rules, found.Rules) || !reflect.DeepEqual(regular.Labels, found.Labels) || !reflect.DeepEqual(regular.Annotations, found.Annotations) {
		found.Rules = regular.Rules
		found.Labels = regular.Labels
		found.Annotations = regular.Annotations
		log.Info("Updating team role", "team", instance.Name, "name", regular.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TeamReconciler) createTeamViewer(instance *tenantv1alpha1.Team) error {
	found := &rbac.ClusterRole{}

	viewer := getTeamViewer(instance.Name)

	if err := controllerutil.SetControllerReference(instance, viewer, r.Scheme); err != nil {
		return err
	}

	err := r.Get(context.TODO(), types.NamespacedName{Name: viewer.Name}, found)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating team role", "team", instance.Name, "name", viewer.Name)
		err = r.Create(context.TODO(), viewer)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		found = viewer
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(viewer.Rules, found.Rules) || !reflect.DeepEqual(viewer.Labels, found.Labels) || !reflect.DeepEqual(viewer.Annotations, found.Annotations) {
		found.Rules = viewer.Rules
		found.Labels = viewer.Labels
		found.Annotations = viewer.Annotations
		log.Info("Updating team role", "team", instance.Name, "name", viewer.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *TeamReconciler) createTeamRoleBindings(instance *tenantv1alpha1.Team) error {
	adminRoleBinding := &rbac.ClusterRoleBinding{}
	adminRoleBinding.Name = getTeamAdminRoleBindingName(instance.Name)
	adminRoleBinding.Labels = map[string]string{constants.TeamLabelKey: instance.Name}
	adminRoleBinding.RoleRef = rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: getTeamAdminRoleName(instance.Name)}

	teamManager := rbac.Subject{APIGroup: "rbac.authorization.k8s.io", Kind: "User", Name: instance.Spec.Manager}

	if teamManager.Name != "" {
		adminRoleBinding.Subjects = []rbac.Subject{teamManager}
	} else {
		adminRoleBinding.Subjects = []rbac.Subject{}
	}

	if err := controllerutil.SetControllerReference(instance, adminRoleBinding, r.Scheme); err != nil {
		return err
	}

	foundRoleBinding := &rbac.ClusterRoleBinding{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: adminRoleBinding.Name}, foundRoleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating team role binding", "team", instance.Name, "name", adminRoleBinding.Name)
		err = r.Create(context.TODO(), adminRoleBinding)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		foundRoleBinding = adminRoleBinding
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(adminRoleBinding.RoleRef, foundRoleBinding.RoleRef) {
		log.Info("Deleting conflict team role binding", "team", instance.Name, "name", adminRoleBinding.Name)
		err = r.Delete(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict team role binding %s, waiting for recreate", foundRoleBinding.Name)
	}

	if teamManager.Name != "" && !hasSubject(foundRoleBinding.Subjects, teamManager) {
		foundRoleBinding.Subjects = append(foundRoleBinding.Subjects, teamManager)
		log.Info("Updating team role binding", "team", instance.Name, "name", adminRoleBinding.Name)
		err = r.Update(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
	}

	regularRoleBinding := &rbac.ClusterRoleBinding{}
	regularRoleBinding.Name = getTeamRegularRoleBindingName(instance.Name)
	regularRoleBinding.Labels = map[string]string{constants.TeamLabelKey: instance.Name}
	regularRoleBinding.RoleRef = rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: getTeamRegularRoleName(instance.Name)}
	regularRoleBinding.Subjects = []rbac.Subject{}

	if err = controllerutil.SetControllerReference(instance, regularRoleBinding, r.Scheme); err != nil {
		return err
	}

	err = r.Get(context.TODO(), types.NamespacedName{Name: regularRoleBinding.Name}, foundRoleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating team role binding", "team", instance.Name, "name", regularRoleBinding.Name)
		err = r.Create(context.TODO(), regularRoleBinding)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		foundRoleBinding = regularRoleBinding
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(regularRoleBinding.RoleRef, foundRoleBinding.RoleRef) {
		log.Info("Deleting conflict team role binding", "team", instance.Name, "name", regularRoleBinding.Name)
		err = r.Delete(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict team role binding %s, waiting for recreate", foundRoleBinding.Name)
	}

	viewerRoleBinding := &rbac.ClusterRoleBinding{}
	viewerRoleBinding.Name = getTeamViewerRoleBindingName(instance.Name)
	viewerRoleBinding.Labels = map[string]string{constants.TeamLabelKey: instance.Name}
	viewerRoleBinding.RoleRef = rbac.RoleRef{APIGroup: "rbac.authorization.k8s.io", Kind: "ClusterRole", Name: getTeamViewerRoleName(instance.Name)}
	viewerRoleBinding.Subjects = []rbac.Subject{}

	if err = controllerutil.SetControllerReference(instance, viewerRoleBinding, r.Scheme); err != nil {
		return err
	}

	err = r.Get(context.TODO(), types.NamespacedName{Name: viewerRoleBinding.Name}, foundRoleBinding)

	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating team role binding", "team", instance.Name, "name", viewerRoleBinding.Name)
		err = r.Create(context.TODO(), viewerRoleBinding)
		// Error reading the object - requeue the request.
		if err != nil {
			return err
		}
		foundRoleBinding = viewerRoleBinding
	} else if err != nil {
		// Error reading the object - requeue the request.
		return err
	}

	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(viewerRoleBinding.RoleRef, foundRoleBinding.RoleRef) {
		log.Info("Deleting conflict team role binding", "team", instance.Name, "name", viewerRoleBinding.Name)
		err = r.Delete(context.TODO(), foundRoleBinding)
		if err != nil {
			return err
		}
		return fmt.Errorf("conflict team role binding %s, waiting for recreate", foundRoleBinding.Name)
	}

	return nil
}

func (r *TeamReconciler) bindNamespaces(instance *tenantv1alpha1.Team) error {

	nsList := &corev1.NamespaceList{}
	options := client.ListOptions{LabelSelector: labels.SelectorFromSet(labels.Set{constants.TeamLabelKey: instance.Name})}
	err := r.List(context.TODO(), nsList, &options)

	if err != nil {
		log.Error(err, fmt.Sprintf("list team %s namespace failed", instance.Name))
		return err
	}

	for _, namespace := range nsList.Items {
		if !metav1.IsControlledBy(&namespace, instance) {
			if err := controllerutil.SetControllerReference(instance, &namespace, r.Scheme); err != nil {
				return err
			}
			log.Info("Bind team", "namespace", namespace.Name, "team", instance.Name)
			err = r.Update(context.TODO(), &namespace)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func hasSubject(subjects []rbac.Subject, user rbac.Subject) bool {
	for _, subject := range subjects {
		if reflect.DeepEqual(subject, user) {
			return true
		}
	}
	return false
}

func getTeamAdmin(teamName string) *rbac.ClusterRole {
	admin := &rbac.ClusterRole{}
	admin.Name = getTeamAdminRoleName(teamName)
	admin.Labels = map[string]string{constants.TeamLabelKey: teamName}
	admin.Annotations = map[string]string{constants.DisplayNameAnnotationKey: constants.TeamAdmin, constants.DescriptionAnnotationKey: teamAdminDescription, constants.CreatorAnnotationKey: constants.System}
	admin.Rules = []rbac.PolicyRule{
		{
			Verbs:         []string{"*"},
			APIGroups:     []string{"*"},
			ResourceNames: []string{teamName},
			Resources:     []string{"teams", "teams/*"},
		},
		//{
		//	Verbs:     []string{"list"},
		//	APIGroups: []string{"iam.kubesphere.io"},
		//	Resources: []string{"users"},
		//},
		//{
		//	Verbs:     []string{"*"},
		//	APIGroups: []string{"openpitrix.io"},
		//	Resources: []string{"applications", "apps", "apps/versions", "apps/events", "apps/action", "apps/audits", "repos", "repos/action", "categories", "attachments"},
		//},
	}

	return admin
}
func getTeamRegular(teamName string) *rbac.ClusterRole {
	regular := &rbac.ClusterRole{}
	regular.Name = getTeamRegularRoleName(teamName)
	regular.Labels = map[string]string{constants.TeamLabelKey: teamName}
	regular.Annotations = map[string]string{constants.DisplayNameAnnotationKey: constants.TeamRegular, constants.DescriptionAnnotationKey: teamRegularDescription, constants.CreatorAnnotationKey: constants.System}
	regular.Rules = []rbac.PolicyRule{
		{
			Verbs:         []string{"get"},
			APIGroups:     []string{"*"},
			Resources:     []string{"teams"},
			ResourceNames: []string{teamName},
		}, {
			Verbs:         []string{"create"},
			APIGroups:     []string{"tenant.kubenebula.io"},
			Resources:     []string{"teams/namespaces"},
			ResourceNames: []string{teamName},
		},
		//{
		//	Verbs:         []string{"get"},
		//	APIGroups:     []string{"iam.kubesphere.io"},
		//	ResourceNames: []string{teamName},
		//	Resources:     []string{"teams/members"},
		//},
		//{
		//	Verbs:     []string{"get", "list"},
		//	APIGroups: []string{"openpitrix.io"},
		//	Resources: []string{"apps/events", "apps/action", "apps/audits"},
		//},
		//
		//{
		//	Verbs:     []string{"*"},
		//	APIGroups: []string{"openpitrix.io"},
		//	Resources: []string{"applications", "apps", "apps/versions", "repos", "repos/action", "categories", "attachments"},
		//},
	}
	return regular
}

func getTeamViewer(teamName string) *rbac.ClusterRole {
	viewer := &rbac.ClusterRole{}
	viewer.Name = getTeamViewerRoleName(teamName)
	viewer.Labels = map[string]string{constants.TeamLabelKey: teamName}
	viewer.Annotations = map[string]string{constants.DisplayNameAnnotationKey: constants.TeamViewer, constants.DescriptionAnnotationKey: teamViewerDescription, constants.CreatorAnnotationKey: constants.System}
	viewer.Rules = []rbac.PolicyRule{
		{
			Verbs:         []string{"get", "list"},
			APIGroups:     []string{"*"},
			ResourceNames: []string{teamName},
			Resources:     []string{"teams", "teams/*"},
		},
		//{
		//	Verbs:     []string{"get", "list"},
		//	APIGroups: []string{"openpitrix.io"},
		//	Resources: []string{"applications", "apps", "apps/versions", "repos", "categories", "attachments"},
		//},
	}
	return viewer
}
func getTeamAdminRoleName(teamName string) string {
	return fmt.Sprintf("team:%s:admin", teamName)
}
func getTeamRegularRoleName(teamName string) string {
	return fmt.Sprintf("team:%s:regular", teamName)
}
func getTeamViewerRoleName(teamName string) string {
	return fmt.Sprintf("team:%s:viewer", teamName)
}
func getTeamAdminRoleBindingName(teamName string) string {
	return fmt.Sprintf("team:%s:admin", teamName)
}

func getTeamRegularRoleBindingName(teamName string) string {
	return fmt.Sprintf("team:%s:regular", teamName)
}

func getTeamViewerRoleBindingName(teamName string) string {
	return fmt.Sprintf("team:%s:viewer", teamName)
}
