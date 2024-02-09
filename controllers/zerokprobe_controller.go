package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/handler"
	zkLogger "github.com/zerok-ai/zk-utils-go/logs"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

// ZerokProbeReconciler reconciles a ZerokProbe object
type ZerokProbeReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ZkCRDProbeHandler *handler.ZkCRDProbeHandler
	Recorder          record.EventRecorder
}

var (
	zerokProbeFinalizerName = "operator.zerok.ai/finalizer"
)

const zerokProbeHandlerLogTag = "ZerokProbeHandler"

//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokprobes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokprobes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokprobes/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *ZerokProbeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	zkLogger.Info(zerokProbeHandlerLogTag, "Reconciling CRD Probe : ")

	zerokProbe := &operatorv1alpha1.ZerokProbe{}
	err := r.Get(ctx, req.NamespacedName, zerokProbe)
	// Write an event to the ContainerSet instance with the namespace and name of the
	// created deployment
	r.Recorder.Event(zerokProbe, "Normal", "ZerokProbeReconciling", fmt.Sprintf("Zerok Probe Reconcile Event."))

	if err != nil {
		// if the resource is not found, then just return (might look useless as this usually happens in case of Delete events)
		zkLogger.Info(zerokProbeHandlerLogTag, "Error occurred while fetching the zerok probe resource, probe might be deleted")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Reconcile logic for each CRD event
	_, err = r.reconcileZerokProbeResource(ctx, zerokProbe, req)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, "Failed to reconcile CustomResource ", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
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
		err := r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbePending,
			"ProbeCreateOrUpdateEvent",
			"Probe_Creating_Or_Updating_in_process",
			fmt.Sprintf("Performing finalizer operations for the custom resource: %s ", zerokProbe.Name))
		if err != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the status of the zerok probe resource")
			return ctrl.Result{}, err
		}

		if !controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
			controllerutil.AddFinalizer(zerokProbe, zerokProbeFinalizerName)
			if err := r.Update(ctx, zerokProbe); err != nil {
				zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the zerok probe resource after adding finalizer")
				return ctrl.Result{}, err
			}
			// Let's re-fetch the Probe Custom Resource after update the status
			// so that we have the latest state of the resource on the cluster
			if err := r.Get(ctx, req.NamespacedName, zerokProbe); err != nil {
				return ctrl.Result{}, err
			}
		}
		if zerokProbe.ObjectMeta.UID == "" {
			// probe is being created
			err = r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbePending, "ProbeCreating", "Probe_Creating", fmt.Sprintf("Started Probe Creation Process : %s", zerokProbe.Spec.Title))
			if err != nil {
				zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the status of the zerok probe resource in creating process")
				return ctrl.Result{}, err
			}
			return r.handleProbeCreation(ctx, req, zerokProbe)
		}

		err = r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbePending, "ProbeUpdating", "Probe_Updating", fmt.Sprintf("Started Probe Updating Process : %s", zerokProbe.Spec.Title))
		if err != nil {
			return ctrl.Result{}, err
		}
		// probe is being updated
		return r.handleProbeUpdate(ctx, req, zerokProbe)
	} else {

		// The object is being deleted

		// Let's add here status "Downgrade" to define that this resource begin its process to be terminated.
		err := r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeDeleting, "ProbeDeleting", "Probe_Deleting", fmt.Sprintf("Started Probe Deleting Process : %s", zerokProbe.Spec.Title))
		if err != nil {
			return ctrl.Result{}, err
		}

		if controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.handleProbeDeletion(ctx, req, zerokProbe); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{RequeueAfter: time.Second * 5}, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(zerokProbe, zerokProbeFinalizerName)
			if err := r.Update(ctx, zerokProbe); err != nil {
				return ctrl.Result{RequeueAfter: time.Second * 5}, err
			}
			// Let's re-fetch the Probe Custom Resource after update the status
			// so that we have the latest state of the resource on the cluster
			if err := r.Get(ctx, req.NamespacedName, zerokProbe); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
}

// handleCreation handles the creation of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeCreation(ctx context.Context, req ctrl.Request, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {

	_, err := r.ZkCRDProbeHandler.CreateCRDProbe(zerokProbe)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileCreating", fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		errStatus := r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeFailed, "ProbeCreationFailed", "Probe_Creation_Failed", fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		if errStatus != nil {
			return ctrl.Result{}, errStatus
		}
		return ctrl.Result{}, err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "CreatedProbe", fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	err = r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeSucceeded, "ProbeCreated", "Probe_Created", fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// handleUpdate handles the update of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeUpdate(ctx context.Context, req ctrl.Request, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {
	_, err := r.ZkCRDProbeHandler.UpdateCRDProbe(zerokProbe)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		errStatus := r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeFailed, "ProbeUpdateFailed", "Probe_Update_Failed", fmt.Sprintf("Error While Updating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		if errStatus != nil {
			return ctrl.Result{}, errStatus
		}
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileUpdating", fmt.Sprintf("Error While Updating CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return ctrl.Result{}, err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Updated Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "UpdatedCRD", fmt.Sprintf("Successfully Updated CRD: %s", zerokProbe.Spec.Title))
	err = r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeSucceeded, "ProbeUpdated", "Probe_Updated", fmt.Sprintf("Successfully Updated Probe: %s", zerokProbe.Spec.Title))
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// handleDeletion handles the deletion of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeDeletion(ctx context.Context, req ctrl.Request, zerokProbe *operatorv1alpha1.ZerokProbe) error {
	zerokProbeVersion := zerokProbe.GetUID()
	_, err := r.ZkCRDProbeHandler.DeleteCRDProbe(string(zerokProbeVersion))
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Deleting Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		errStatus := r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeFailed, "ProbeDeletionFailed", "Probe_Deletion_Failed", fmt.Sprintf("Error While Deleting Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		if errStatus != nil {
			return err
		}
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileDeleting", fmt.Sprintf("Error While Deleting CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Deleted Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "DeletedCRD", fmt.Sprintf("Successfully Deleted CRD: %s", zerokProbe.Spec.Title))
	err = r.UpdateProbeResourceStatus(ctx, req, zerokProbe, operatorv1alpha1.ProbeSucceeded, "ProbeDeleted", "Probe_Deleted", fmt.Sprintf("Successfully Deleted Probe: %s", zerokProbe.Spec.Title))
	if err != nil {
		return err
	}
	return nil
}

func (r *ZerokProbeReconciler) UpdateProbeResourceStatus(ctx context.Context, req ctrl.Request, zerokProbe *operatorv1alpha1.ZerokProbe, probeStatus operatorv1alpha1.ZerokProbePhase, probeStatusType string, probeStatusReason string, probeStatusMessage string) error {
	meta.SetStatusCondition(&zerokProbe.Status.Conditions, metav1.Condition{Type: probeStatusType,
		Status: metav1.ConditionTrue, Reason: probeStatusReason,
		Message: probeStatusMessage})
	zerokProbe.Status.Phase = probeStatus

	if err := r.Status().Update(ctx, zerokProbe); err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe Status: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return err
	}

	// Let's re-fetch the Probe Custom Resource after update the status
	// so that we have the latest state of the resource on the cluster
	if err := r.Get(ctx, req.NamespacedName, zerokProbe); err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while fetching the zerok probe resource after updating the status in creating or updating process")
		return err
	}

	return nil
}
