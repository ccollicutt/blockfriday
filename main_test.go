package main

import (
	"encoding/json"
	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"k8s.io/apimachinery/pkg/runtime"
)

type mockClient struct{}

func TestHandleAdmissionRequest(t *testing.T) {
	t.Run("it denies requests on Friday", func(t *testing.T) {
		// Set up the test by creating an AdmissionReview for a Deployment
		admissionReview := createAdmissionReview()

		// Create a mock Kubernetes client
		client := &mockClient{}

		// Override the isTodayFriday function to always return true
		isTodayFriday := func() bool {
			return true
		}

		// Call the function under test
		response := handleAdmissionRequest(admissionReview, client, isTodayFriday)

		// Check the result
		if response.Allowed {
			t.Errorf("Expected handleAdmissionRequest to deny the request, but it was allowed.")
		}
	})

	t.Run("it allows requests not on Friday", func(t *testing.T) {
		// Set up the test by creating an AdmissionReview for a Deployment
		admissionReview := createAdmissionReview()

		// Create a mock Kubernetes client
		client := &mockClient{}

		// Override the isTodayFriday function to always return false
		isTodayFriday := func() bool {
			return false
		}

		// Call the function under test
		response := handleAdmissionRequest(admissionReview, client, isTodayFriday)

		// Check the result
		if !response.Allowed {
			t.Errorf("Expected handleAdmissionRequest to allow the request, but it was denied.")
		}
	})
}

func createAdmissionReview() v1.AdmissionReview {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "test-namespace",
		},
	}
	deploymentJSON, _ := json.Marshal(deployment)

	return v1.AdmissionReview{
		Request: &v1.AdmissionRequest{
			UID: "test-uid",
			Object: runtime.RawExtension{
				Raw: deploymentJSON,
			},
		},
	}
}
