package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/handler"
	zkLogger "github.com/zerok-ai/zk-utils-go/logs"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
	probeStatusType         = "ProbeStatus"
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
	//r.Recorder.Event(zerokProbe, "Normal", "ZerokProbeReconciling", fmt.Sprintf("Zerok Probe Reconcile Event."))

	if err != nil {
		// if the resource is not found, then just return (might look useless as this usually happens in case of Delete events)
		zkLogger.Info(zerokProbeHandlerLogTag, "Error occurred while fetching the zerok probe resource might be deleted.")
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

		if zerokProbe.ObjectMeta.UID == "" {
			// probe create scenario
			err := r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Creating", fmt.Sprintf("Started Probe Creation Process : %s", zerokProbe.Spec.Title))
			if err != nil {
				zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the status of the zerok probe resource in creating process")
				return ctrl.Result{}, err
			}
			//result, err2 := r.addFinalizerIfNotPresent(ctx, zerokProbe)
			//if err2 != nil {
			//	return result, err2
			//}
			return r.handleProbeCreation(ctx, zerokProbe)
		}

		//Probe update scenario
		err := r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Updating", fmt.Sprintf("Started Probe Updating Process : %s", zerokProbe.Spec.Title))
		if err != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the status of the zerok probe resource in updating process")
			return ctrl.Result{}, err
		}
		// probe is being updated
		return r.handleProbeUpdate(ctx, zerokProbe)
	} else {

		// The object is being deleted

		// Let's add here status "Downgrade" to define that this resource begin its process to be terminated.
		err := r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Deleting", fmt.Sprintf("Started Probe Deleting Process : %s", zerokProbe.Spec.Title))
		if err != nil {
			return ctrl.Result{}, err
		}

		//if controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
		// our finalizer is present, so lets handle any external dependency
		if err := r.handleProbeDeletion(ctx, zerokProbe); err != nil {
			// if fail to delete the external dependency here, return with error
			// so that it can be retried
			return ctrl.Result{RequeueAfter: time.Second * 5}, err
		}

		//// remove our finalizer from the list and update it.
		//controllerutil.RemoveFinalizer(zerokProbe, zerokProbeFinalizerName)
		//if err := r.Update(ctx, zerokProbe); err != nil {
		//	return ctrl.Result{RequeueAfter: time.Second * 5}, err
		//}
		//err := r.FetchUpdatedProbeObject(ctx, zerokProbe.Namespace, zerokProbe.Name, zerokProbe)
		//if err != nil {
		//	return ctrl.Result{}, err
		//}
		//}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
}

func (r *ZerokProbeReconciler) addFinalizerIfNotPresent(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
		controllerutil.AddFinalizer(zerokProbe, zerokProbeFinalizerName)
		if err := r.Update(ctx, zerokProbe); err != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the zerok probe resource after adding finalizer")
			return ctrl.Result{}, err
		}

		err := r.FetchUpdatedProbeObject(ctx, zerokProbe.Namespace, zerokProbe.Name, zerokProbe)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// handleCreation handles the creation of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeCreation(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {

	_, err := r.ZkCRDProbeHandler.CreateCRDProbe(zerokProbe)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileCreating", fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		errStatus := r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Creation_Failed", fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		if errStatus != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe Status: %s with error: %s", zerokProbe.Spec.Title, errStatus.Error()))
		}
		return ctrl.Result{}, err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "CreatedProbe", fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	err = r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Created", fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	if err != nil {
		//TODO: Do we have to return error here? Since only status update failed.
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// handleUpdate handles the update of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeUpdate(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {
	_, err := r.ZkCRDProbeHandler.UpdateCRDProbe(zerokProbe)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		errStatus := r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Update_Failed", fmt.Sprintf("Error While Updating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		if errStatus != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe Status: %s with error: %s", zerokProbe.Spec.Title, errStatus.Error()))
		}
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileUpdating", fmt.Sprintf("Error While Updating CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return ctrl.Result{}, err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Updated Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "UpdatedCRD", fmt.Sprintf("Successfully Updated CRD: %s", zerokProbe.Spec.Title))
	err = r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Updated", fmt.Sprintf("Successfully Updated Probe: %s", zerokProbe.Spec.Title))
	if err != nil {
		//TODO: Do we have to return error here? Since only status update failed.
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// handleDeletion handles the deletion of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeDeletion(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) error {
	zerokProbeVersion := zerokProbe.GetUID()
	_, err := r.ZkCRDProbeHandler.DeleteCRDProbe(string(zerokProbeVersion))
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Deleting Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		errStatus := r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Deletion_Failed", fmt.Sprintf("Error While Deleting Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		if errStatus != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe Status: %s with error: %s", zerokProbe.Spec.Title, errStatus.Error()))
		}
		//r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileDeleting", fmt.Sprintf("Error While Deleting CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Deleted Probe: %s", zerokProbe.Spec.Title))
	//r.Recorder.Event(zerokProbe, "Normal", "DeletedCRD", fmt.Sprintf("Successfully Deleted CRD: %s", zerokProbe.Spec.Title))
	err = r.UpdateProbeResourceStatus(ctx, zerokProbe, probeStatusType, "Probe_Deleted", fmt.Sprintf("Successfully Deleted Probe: %s", zerokProbe.Spec.Title))
	if err != nil {
		return err
	}
	return nil
}

func (r *ZerokProbeReconciler) UpdateProbeResourceStatus(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe, probeStatusType string, probeStatusReason string, probeStatusMessage string) error {
	//meta.SetStatusCondition(&zerokProbe.Status.Conditions, metav1.Condition{Type: probeStatusType,
	//	Status: metav1.ConditionTrue, Reason: probeStatusReason,
	//	Message: probeStatusMessage})
	//if err := r.Status().Update(ctx, zerokProbe); err != nil {
	//	zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe Status: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
	//	return err
	//}
	//return r.FetchUpdatedProbeObject(ctx, zerokProbe.Namespace, zerokProbe.Name, zerokProbe)
	return nil
}

// Let's re-fetch the Probe Custom Resource after update the status
// so that we have the latest state of the resource on the cluster
func (r *ZerokProbeReconciler) FetchUpdatedProbeObject(ctx context.Context, namespace, name string, zerokProbe *operatorv1alpha1.ZerokProbe) error {
	if err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, zerokProbe); err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while fetching the zerok probe resource")
		return err
	}
	return nil
}
