package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/handler"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ZerokProbeReconciler reconciles a ZerokProbe object
type ZerokProbeReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ZkCRDProbeHandler *handler.ZkCRDProbeHandler
	Recorder          record.EventRecorder
}

var (
	finalizers              []string = []string{"finalizers.operator.zerok.ai"}
	zerokProbeFinalizerName          = "operator.zerok.ai/finalizer"
)

//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokprobes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokprobes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokprobes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ZerokProbe object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ZerokProbeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CRD Probe : ")

	zerokProbe := &operatorv1alpha1.ZerokProbe{}
	err := r.Get(ctx, req.NamespacedName, zerokProbe)
	// Write an event to the ContainerSet instance with the namespace and name of the
	// created deployment
	r.Recorder.Event(zerokProbe, "Normal", "Created", fmt.Sprintf(""))

	// finalizers for deleting the probe

	//TODO :: move below logic to respective handlers for creation, deletion, update bypassing is NOT FOUND error for deletion events
	if err != nil {
		// if the resource is not found, then just return (might look useless as this usually happens in case of Delete events)
		logger.Error(err, "Error occurred while fetching the zerok probe resource")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Let's just set the status as Unknown when no status are available
	if zerokProbe.Status.Conditions == nil || len(zerokProbe.Status.Conditions) == 0 {
		meta.SetStatusCondition(&zerokProbe.Status.Conditions, metav1.Condition{Type: "", Status: metav1.ConditionUnknown, Reason: "Reconciling", Message: "Starting reconciliation"})
		zerokProbe.Status.Phase = operatorv1alpha1.ProbeUnknown
		if err = r.Status().Update(ctx, zerokProbe); err != nil {

			return ctrl.Result{}, err
		}
		// Let's re-fetch the Probe Custom Resource after update the status
		// so that we have the latest state of the resource on the cluster
		if err := r.Get(ctx, req.NamespacedName, zerokProbe); err != nil {
			//log.Error(err, "Failed to re-fetch memcached")
			return ctrl.Result{}, err
		}
	}

	// Reconcile logic for each CRD event
	_, err = r.reconcileZerokProbeResource(ctx, zerokProbe, req)
	if err != nil {
		logger.Error(err, "Failed to reconcile CustomResource")
		return ctrl.Result{}, err
	}

	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZerokProbeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.ZerokProbe{}).
		Complete(r)
}

func (r *ZerokProbeReconciler) reconcileZerokProbeResource(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe, req ctrl.Request) (ctrl.Result, error) {

	// check if it is deletion
	// examine DeletionTimestamp to determine if object is under deletion
	if zerokProbe.ObjectMeta.GetDeletionTimestamp().IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		// Let's add here status "Downgrade" to define that this resource begin its process to be terminated.
		meta.SetStatusCondition(&zerokProbe.Status.Conditions, metav1.Condition{Type: "",
			Status: metav1.ConditionUnknown, Reason: "Finalizing",
			Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", zerokProbe.Name)})
		zerokProbe.Status.Phase = operatorv1alpha1.ProbeRunning
		if err := r.Status().Update(ctx, zerokProbe); err != nil {
			return ctrl.Result{}, err
		}

		// Let's re-fetch the Probe Custom Resource after update the status
		// so that we have the latest state of the resource on the cluster
		if err := r.Get(ctx, req.NamespacedName, zerokProbe); err != nil {
			return ctrl.Result{}, err
		}

		if !controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
			controllerutil.AddFinalizer(zerokProbe, zerokProbeFinalizerName)
			if err := r.Update(ctx, zerokProbe); err != nil {
				return ctrl.Result{}, err
			}
		}
		if zerokProbe.ObjectMeta.UID == "" {
			// probe is being created
			return r.handleProbeCreation(ctx, zerokProbe)
		}
		// probe is being updated
		return r.handleProbeUpdate(ctx, zerokProbe)
	} else {

		// The object is being deleted

		// Let's add here status "Downgrade" to define that this resource begin its process to be terminated.
		meta.SetStatusCondition(&zerokProbe.Status.Conditions, metav1.Condition{Type: "",
			Status: metav1.ConditionUnknown, Reason: "Finalizing",
			Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", zerokProbe.Name)})
		zerokProbe.Status.Phase = operatorv1alpha1.ProbeDeleting
		if err := r.Status().Update(ctx, zerokProbe); err != nil {
			return ctrl.Result{}, err
		}

		if controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.handleProbeDeletion(ctx, zerokProbe); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{Requeue: true}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(zerokProbe, zerokProbeFinalizerName)
			if err := r.Update(ctx, zerokProbe); err != nil {
				return ctrl.Result{Requeue: true}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
}

// handleCreation handles the creation of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeCreation(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {

	_, err := r.ZkCRDProbeHandler.CreateCRDProbe(zerokProbe)
	if err != nil {
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileCreating", fmt.Sprintf("Error While Creating CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return ctrl.Result{Requeue: true}, err
	}

	r.Recorder.Event(zerokProbe, "Normal", "CreatedCRD", fmt.Sprintf("Successfully Created CRD: %s", zerokProbe.Spec.Title))

	// Let's add here status "Downgrade" to define that this resource begin its process to be terminated.
	meta.SetStatusCondition(&zerokProbe.Status.Conditions, metav1.Condition{Type: "",
		Status: metav1.ConditionUnknown, Reason: "Finalizing",
		Message: fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", zerokProbe.Name)})
	zerokProbe.Status.Phase = operatorv1alpha1.ProbeSucceeded
	if err := r.Status().Update(ctx, zerokProbe); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handleUpdate handles the update of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeUpdate(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {
	_, err := r.ZkCRDProbeHandler.UpdateCRDProbe(zerokProbe)
	if err != nil {
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileUpdating", fmt.Sprintf("Error While Updating CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return ctrl.Result{Requeue: true}, err
	}

	r.Recorder.Event(zerokProbe, "Normal", "UpdatedCRD", fmt.Sprintf("Successfully Updated CRD: %s", zerokProbe.Spec.Title))
	return ctrl.Result{Requeue: true}, nil
}

// handleDeletion handles the deletion of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeDeletion(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) error {
	zerokProbeVersion := zerokProbe.GetUID()
	fmt.Print(zerokProbeVersion)
	_, err := r.ZkCRDProbeHandler.DeleteCRDProbe(string(zerokProbeVersion))
	if err != nil {
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileDeleting", fmt.Sprintf("Error While Deleting CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return err
	}

	r.Recorder.Event(zerokProbe, "Normal", "DeletedCRD", fmt.Sprintf("Successfully Deleted CRD: %s", zerokProbe.Spec.Title))
	return nil
}
