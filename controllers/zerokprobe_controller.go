package controllers

import (
	"context"
	"fmt"
	operatorv1alpha1 "github.com/zerok-ai/zk-operator/api/v1alpha1"
	"github.com/zerok-ai/zk-operator/internal/handler"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ZerokCrdReconciler reconciles a ZerokCrd object
type ZerokCrdReconciler struct {
	client.Client
	Scheme            *runtime.Scheme
	ZkCRDProbeHandler *handler.ZkCRDProbeHandler
}

var (
	finalizers []string = []string{"finalizers.operator.zerok.ai"}
)

//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokcrds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokcrds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.zerok.ai.zerok.ai,resources=zerokcrds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ZerokProbe object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ZerokCrdReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling CRD Probe : ")

	zerokCrd := &operatorv1alpha1.ZerokCrd{}
	err := r.Get(ctx, req.NamespacedName, zerokCrd)

	//TODO :: move below logic to respective handlers for creation, deletion, update bypassing is NOT FOUND error for deletion events
	if err != nil && !errors.IsNotFound(err) {
		// if the resource is not found, then just return (might look useless as this usually happens in case of Delete events)
		logger.Error(err, "Error occurred while fetching the zerok probe resource")
		return ctrl.Result{}, err
	}

	// Reconcile logic for each CRD event
	err = r.reconcileZerokProbeResource(ctx, zerokCrd, req)
	if err != nil {
		logger.Error(err, "Failed to reconcile CustomResource")
		return ctrl.Result{}, err
	}

	//status := operatorv1alpha1.ZerokProbeStatus{
	//	IsCreated: true,
	//}

	//if !reflect.DeepEqual(zerokCrd.Status, status) {
	//	zerokCrd.Status = status
	//	err := r.Client.Status().Update(ctx, zerokCrd)
	//	if err != nil {
	//		logger.Error(err, "Error occurred while updating the probe resource")
	//		return reconcile.Result{}, err
	//	}
	//}
	return ctrl.Result{Requeue: true}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ZerokCrdReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&operatorv1alpha1.ZerokCrd{}).
		Complete(r)
}

func (r *ZerokCrdReconciler) reconcileZerokProbeResource(ctx context.Context, zerokCrd *operatorv1alpha1.ZerokCrd, req ctrl.Request) error {

	// check if it is deletion
	if !zerokCrd.ObjectMeta.GetDeletionTimestamp().IsZero() {
		err := r.handleProbeDeletion(ctx, zerokCrd)
		if err != nil {
			return err
		}
	} else {
		if zerokCrd.ObjectMeta.UID == "" {
			// probe is being created
			return r.handleProbeCreation(ctx, zerokCrd)
		}
		// probe is being updated
		return r.handleProbeUpdate(ctx, zerokCrd)
	}

	return nil
}

// handleCreation handles the creation of the ZerokProbe
func (r *ZerokCrdReconciler) handleProbeCreation(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokCrd) error {

	_, err := r.ZkCRDProbeHandler.CreateCRDProbe(zerokProbe)
	if err != nil {
		return err
	}

	return nil
}

// handleUpdate handles the update of the ZerokProbe
func (r *ZerokCrdReconciler) handleProbeUpdate(ctx context.Context, zerokProbe *operatorv1alpha1.ZerokCrd) error {
	_, err := r.ZkCRDProbeHandler.UpdateCRDProbe(zerokProbe)
	if err != nil {
		return err
	}
	return nil
}

// handleDeletion handles the deletion of the ZerokProbe
func (r *ZerokCrdReconciler) handleProbeDeletion(ctx context.Context, zerokCrd *operatorv1alpha1.ZerokCrd) error {
	zerokProbeVersion := zerokCrd.GetUID()
	fmt.Print(zerokProbeVersion)
	_, err := r.ZkCRDProbeHandler.DeleteCRDProbe(string(zerokProbeVersion))
	if err != nil {
		return err
	}
	return nil
}
