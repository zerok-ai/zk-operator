package webhook

import (
	"bytes"
	"context"
	"fmt"
	"github.com/zerok-ai/zk-operator/internal/utils"
	"reflect"

	"github.com/zerok-ai/zk-operator/internal/config"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WebhookHandler struct {
	webhookConfig config.WebhookConfig
	caPem         *bytes.Buffer
	killed        bool
}

func (h *WebhookHandler) Init(caPEM *bytes.Buffer, config config.WebhookConfig) {
	h.caPem = caPEM
	h.webhookConfig = config
	err := h.CreateOrUpdateMutatingWebhookConfiguration()
	if err != nil {
		msg := fmt.Sprintf("Failed to create or update the mutating webhook configuration: %v. Stopping initialization of the pod.\n", err)
		fmt.Println(msg)
		return
	}
}

func (h *WebhookHandler) deleteMutatingWebhookConfiguration() error {
	clientset, err := utils.GetK8sClient()
	if err != nil {
		return err
	}
	mutatingWebhookConfigV1Client := clientset.AdmissionregistrationV1()
	err = mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Delete(context.TODO(), h.webhookConfig.Name, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Error while deleting operator webhook %v.", err)
		return err
	}
	return nil
}

func (h *WebhookHandler) CreateOrUpdateMutatingWebhookConfiguration() error {

	clientset, err := utils.GetK8sClient()
	if err != nil {
		return err
	}

	mutatingWebhookConfigV1Client := clientset.AdmissionregistrationV1()

	fmt.Printf("Creating or updating the mutatingwebhookconfiguration\n")

	ignore := admissionregistrationv1.Ignore
	sideEffect := admissionregistrationv1.SideEffectClassNone
	mutatingWebhookConfig := h.createMutatingWebhookConfig(sideEffect, h.caPem, ignore)

	existingWebhookConfig, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Get(context.TODO(), h.webhookConfig.Name, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {

		//Scenario where there is not existing webhook. So we are creating a new webhook.
		if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Create(context.TODO(), mutatingWebhookConfig, metav1.CreateOptions{}); err != nil {
			fmt.Printf("Failed to create the mutatingwebhookconfiguration: %s\n", h.webhookConfig.Name)
			return err
		}
		fmt.Printf("Created mutatingwebhookconfiguration: %s\n", h.webhookConfig.Name)

	} else if err != nil {

		//Scenario where we failed to check if there was any existing webhook.
		fmt.Printf("Failed to check the mutatingwebhookconfiguration: %s\n", h.webhookConfig.Name)
		fmt.Printf("The error is %v\n", err.Error())
		return err

	} else if !areWebHookConfigsSame(existingWebhookConfig, mutatingWebhookConfig) {

		//Scenario where we have to update the existing webhook.
		mutatingWebhookConfig.ObjectMeta.ResourceVersion = existingWebhookConfig.ObjectMeta.ResourceVersion
		if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Update(context.TODO(), mutatingWebhookConfig, metav1.UpdateOptions{}); err != nil {
			fmt.Printf("Failed to update the mutatingwebhookconfiguration: %s", h.webhookConfig.Name)
			return err
		}
		fmt.Printf("Updated the mutatingwebhookconfiguration: %s\n", h.webhookConfig.Name)

	} else {

		//Scenario where there is no need to update the existing webhook.
		fmt.Printf("The mutatingwebhookconfiguration: %s already exists and has no change\n", h.webhookConfig.Name)

	}

	return nil
}

func (h *WebhookHandler) createMutatingWebhookConfig(sideEffect admissionregistrationv1.SideEffectClass, caPEM *bytes.Buffer, ignore admissionregistrationv1.FailurePolicyType) *admissionregistrationv1.MutatingWebhookConfiguration {
	mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: h.webhookConfig.Name,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    "zk-webhook.zerok.ai",
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caPEM.Bytes(),
				Service: &admissionregistrationv1.ServiceReference{
					Name:      h.webhookConfig.Service,
					Namespace: h.webhookConfig.Namespace,
					Path:      &h.webhookConfig.Path,
				},
			},
			Rules: []admissionregistrationv1.RuleWithOperations{
				{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
						admissionregistrationv1.Update,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				},
			},
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"zk-injection": "enabled",
				},
			},
			FailurePolicy: &ignore,
		}},
	}
	return mutatingWebhookConfig
}

func areWebHookConfigsSame(foundWebhookConfig *admissionregistrationv1.MutatingWebhookConfiguration, mutatingWebhookConfig *admissionregistrationv1.MutatingWebhookConfiguration) bool {
	if len(foundWebhookConfig.Webhooks) != len(mutatingWebhookConfig.Webhooks) {
		return false
	}

	for i, foundWebhookConfig := range foundWebhookConfig.Webhooks {
		mutatingWebhookConfig := mutatingWebhookConfig.Webhooks[i]
		equal := foundWebhookConfig.Name == mutatingWebhookConfig.Name &&
			reflect.DeepEqual(foundWebhookConfig.AdmissionReviewVersions, mutatingWebhookConfig.AdmissionReviewVersions) &&
			reflect.DeepEqual(foundWebhookConfig.SideEffects, mutatingWebhookConfig.SideEffects) &&
			reflect.DeepEqual(foundWebhookConfig.FailurePolicy, mutatingWebhookConfig.FailurePolicy) &&
			reflect.DeepEqual(foundWebhookConfig.Rules, mutatingWebhookConfig.Rules) &&
			reflect.DeepEqual(foundWebhookConfig.NamespaceSelector, mutatingWebhookConfig.NamespaceSelector) &&
			reflect.DeepEqual(foundWebhookConfig.ClientConfig.CABundle, mutatingWebhookConfig.ClientConfig.CABundle) &&
			reflect.DeepEqual(foundWebhookConfig.ClientConfig.Service, mutatingWebhookConfig.ClientConfig.Service)
		if !equal {
			return false
		}
	}
	return true
}

func (h *WebhookHandler) CleanUpOnkill() error {
	fmt.Printf("Kill method in webhook.\n")
	return h.deleteMutatingWebhookConfiguration()
}
