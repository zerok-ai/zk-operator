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

	if err != nil {
		// if the resource is not found, then just return (might look useless as this usually happens in case of Delete events)
		zkLogger.Info(zerokProbeHandlerLogTag, "Error occurred while fetching the zerok probe resource might be deleted.")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Recorder.Event(zerokProbe, "Normal", "ZerokProbeReconciling", fmt.Sprintf("Zerok Probe Reconcile Event."))

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
		// The object is not being deleted

		if zerokProbe.ObjectMeta.UID == "" || zerokProbe.ObjectMeta.Finalizers == nil || len(zerokProbe.ObjectMeta.Finalizers) == 0 {
			// probe create scenario
			r.Recorder.Event(zerokProbe, "Normal", "CreatingProbe", fmt.Sprintf("Started Probe Creation Process : %s", zerokProbe.Spec.Title))
			//so if it does not have our finalizer,
			// then lets add the finalizer and update the object. This is equivalent
			// registering our finalizer.
			creationResult, err := r.handleProbeCreation(ctx, zerokProbe)
			if err != nil {
				return creationResult, err
			}

			//add only finalizer if creation is successful and no error occurred
			err = r.addFinalizerIfNotPresent(ctx, zerokProbe)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}

		// probe is being updated
		return r.handleProbeUpdate(ctx, zerokProbe)
	} else {
		// The object is being deleted
		// Let's add here status "Downgrade" to define that this resource begin its process to be terminated.
		if controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.handleProbeDeletion(ctx, zerokProbe); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				//reconciliation after 5 seconds to retry
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(zerokProbe, zerokProbeFinalizerName)
			if err := r.Update(ctx, zerokProbe); err != nil {
				//reconciliation after 5 seconds to retry
				return ctrl.Result{RequeueAfter: time.Second * 5}, nil
			}
			//TODO:: can be removed
			err := r.FetchUpdatedProbeObject(ctx, zerokProbe.Namespace, zerokProbe.Name, zerokProbe)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}
}

func (r *ZerokProbeReconciler) addFinalizerIfNotPresent(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) error {
	if !controllerutil.ContainsFinalizer(zerokProbe, zerokProbeFinalizerName) {

		err := r.FetchUpdatedProbeObject(ctx, zerokProbe.Namespace, zerokProbe.Name, zerokProbe)
		if err != nil {
			return err
		}

		zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Adding Finalizer to the ZerokProbe: %s", zerokProbe.Spec.Title))
		controllerutil.AddFinalizer(zerokProbe, zerokProbeFinalizerName)

		if err = r.Update(ctx, zerokProbe); err != nil {
			zkLogger.Error(zerokProbeHandlerLogTag, "Error occurred while updating the zerok probe resource after adding finalizer")
			return err
		}

		err = r.FetchUpdatedProbeObject(ctx, zerokProbe.Namespace, zerokProbe.Name, zerokProbe)
		if err != nil {
			return err
		}
	}
	return nil
}

// handleCreation handles the creation of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeCreation(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {

	_, err := r.ZkCRDProbeHandler.CreateCRDProbe(zerokProbe)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileCreating", fmt.Sprintf("Error While Creating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return ctrl.Result{}, err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "CreatedProbe", fmt.Sprintf("Successfully Created Probe: %s", zerokProbe.Spec.Title))
	return ctrl.Result{}, nil
}

// handleUpdate handles the update of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeUpdate(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) (ctrl.Result, error) {
	_, err := r.ZkCRDProbeHandler.UpdateCRDProbe(zerokProbe)
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Updating Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		r.Recorder.Event(zerokProbe, "Warning", "ErrorWhileUpdating", fmt.Sprintf("Error While Updating CRD: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return ctrl.Result{}, err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Updated Probe: %s", zerokProbe.Spec.Title))
	r.Recorder.Event(zerokProbe, "Normal", "UpdatedCRD", fmt.Sprintf("Successfully Updated CRD: %s", zerokProbe.Spec.Title))
	return ctrl.Result{}, nil
}

// handleDeletion handles the deletion of the ZerokProbe
func (r *ZerokProbeReconciler) handleProbeDeletion(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokProbe) error {
	zerokProbeVersion := zerokProbe.GetUID()
	_, err := r.ZkCRDProbeHandler.DeleteCRDProbe(string(zerokProbeVersion))
	if err != nil {
		zkLogger.Error(zerokProbeHandlerLogTag, fmt.Sprintf("Error While Deleting Probe: %s with error: %s", zerokProbe.Spec.Title, err.Error()))
		return err
	}

	zkLogger.Info(zerokProbeHandlerLogTag, fmt.Sprintf("Successfully Deleted Probe: %s", zerokProbe.Spec.Title))
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
