package utils

import (
	"bytes"
	"context"
	"fmt"
	"reflect"

	"github.com/zerok-ai/zk-operator/internal/config"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	webhookName = "zk-webhook"
	webhookPath = "/zk-injector"
)

func CreateOrUpdateMutatingWebhookConfiguration(caPEM *bytes.Buffer, cfg config.WebhookConfig) error {

	clientset, err := GetK8sClient()
	if err != nil {
		return err
	}

	mutatingWebhookConfigV1Client := clientset.AdmissionregistrationV1()

	fmt.Printf("Creating or updating the mutatingwebhookconfiguration\n")

	ignore := admissionregistrationv1.Ignore
	sideEffect := admissionregistrationv1.SideEffectClassNone
	mutatingWebhookConfig := createMutatingWebhook(sideEffect, caPEM, cfg.Service, cfg.Namespace, ignore)

	existingWebhookConfig, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Get(context.TODO(), webhookName, metav1.GetOptions{})
	if err != nil && apierrors.IsNotFound(err) {

		//Scenario where there is not existing webhook. So we are creating a new webhook.
		if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Create(context.TODO(), mutatingWebhookConfig, metav1.CreateOptions{}); err != nil {
			fmt.Printf("Failed to create the mutatingwebhookconfiguration: %s\n", webhookName)
			return err
		}
		fmt.Printf("Created mutatingwebhookconfiguration: %s\n", webhookName)

	} else if err != nil {

		//Scenario where we failed to check if there was any existing webhook.
		fmt.Printf("Failed to check the mutatingwebhookconfiguration: %s\n", webhookName)
		fmt.Printf("The error is %v\n", err.Error())
		return err

	} else if !areWebHooksSame(existingWebhookConfig, mutatingWebhookConfig) {

		//Scenario where we have to update the existing webhook.
		mutatingWebhookConfig.ObjectMeta.ResourceVersion = existingWebhookConfig.ObjectMeta.ResourceVersion
		if _, err := mutatingWebhookConfigV1Client.MutatingWebhookConfigurations().Update(context.TODO(), mutatingWebhookConfig, metav1.UpdateOptions{}); err != nil {
			fmt.Printf("Failed to update the mutatingwebhookconfiguration: %s", webhookName)
			return err
		}
		fmt.Printf("Updated the mutatingwebhookconfiguration: %s\n", webhookName)

	} else {

		//Scenario where there is no need to update the existing webhook.
		fmt.Printf("The mutatingwebhookconfiguration: %s already exists and has no change\n", webhookName)

	}

	return nil
}

func createMutatingWebhook(sideEffect admissionregistrationv1.SideEffectClass, caPEM *bytes.Buffer, webhookService string, webhookNamespace string, ignore admissionregistrationv1.FailurePolicyType) *admissionregistrationv1.MutatingWebhookConfiguration {
	mutatingWebhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: webhookName,
		},
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			Name:                    "zk-webhook.zerok.ai",
			AdmissionReviewVersions: []string{"v1"},
			SideEffects:             &sideEffect,
			ClientConfig: admissionregistrationv1.WebhookClientConfig{
				CABundle: caPEM.Bytes(),
				Service: &admissionregistrationv1.ServiceReference{
					Name:      webhookService,
					Namespace: webhookNamespace,
					Path:      &webhookPath,
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

func areWebHooksSame(foundWebhookConfig *admissionregistrationv1.MutatingWebhookConfiguration, mutatingWebhookConfig *admissionregistrationv1.MutatingWebhookConfiguration) bool {
	if len(foundWebhookConfig.Webhooks) != len(mutatingWebhookConfig.Webhooks) {
		return false
	}
	len := len(foundWebhookConfig.Webhooks)
	for i := 0; i < len; i++ {
		equal := foundWebhookConfig.Webhooks[i].Name == mutatingWebhookConfig.Webhooks[i].Name &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].AdmissionReviewVersions, mutatingWebhookConfig.Webhooks[i].AdmissionReviewVersions) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].SideEffects, mutatingWebhookConfig.Webhooks[i].SideEffects) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].FailurePolicy, mutatingWebhookConfig.Webhooks[i].FailurePolicy) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].Rules, mutatingWebhookConfig.Webhooks[i].Rules) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].NamespaceSelector, mutatingWebhookConfig.Webhooks[i].NamespaceSelector) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].ClientConfig.CABundle, mutatingWebhookConfig.Webhooks[i].ClientConfig.CABundle) &&
			reflect.DeepEqual(foundWebhookConfig.Webhooks[i].ClientConfig.Service, mutatingWebhookConfig.Webhooks[i].ClientConfig.Service)
		if !equal {
			return false
		}
	}
	return true
}
