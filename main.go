package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

const (
	admissionWebhookPort = 8080
	tlsKeyFile           = "/certs/tls.key"
	tlsCertFile          = "/certs/tls.crt"
	admissionAPIVersion  = "admission.k8s.io/v1"
	admissionKind        = "AdmissionReview"
)

type KubernetesClient interface {
	// FIXME Include methods used from *kubernetes.Clientset for testing
}

func main() {
	klog.Info("Starting admission controller...")

	// Set up the Kubernetes client
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Error creating Kubernetes client config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating Kubernetes client: %v", err)
	}

	// Set up the HTTP server with the admission handler
	http.HandleFunc("/validate", func(w http.ResponseWriter, request *http.Request) {
		serveAdmissionRequest(w, request, clientset, isFriday)
	})

	// Create a context with cancellation support for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := &http.Server{Addr: ":8443"}
	startServer(server, tlsCertFile, tlsKeyFile)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			klog.Info("Admission controller server is running.")
		case <-signalCh:
			klog.Info("Received shutdown signal. Shutting down...")
			server.Shutdown(ctx)
			return
		}
	}
}

func startServer(server *http.Server, tlsCertFile string, tlsKeyFile string) {
	go func() {
		if err := server.ListenAndServeTLS(tlsCertFile, tlsKeyFile); err != nil {
			klog.Fatalf("Failed to start admission controller: %v", err)
		} else {
			klog.Info("Admission controller started.")
		}
	}()
}

func serveAdmissionRequest(w http.ResponseWriter, request *http.Request, clientset KubernetesClient, isTodayFriday func() bool) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		klog.Errorf("Error reading admission request body: %v", err)
		http.Error(w, "Error reading admission request body", http.StatusBadRequest)
		return
	}
	request.Body.Close()

	// Deserialize the admission request
	admissionReview, err := decodeAdmissionReview(body)
	if err != nil {
		klog.Errorf("Error decoding admission request: %v", err)
		http.Error(w, "Error decoding admission request", http.StatusBadRequest)
		return
	}

	response := handleAdmissionRequest(admissionReview, clientset, isTodayFriday)

	responseReview := v1.AdmissionReview{
		Response: response,
		TypeMeta: metav1.TypeMeta{
			APIVersion: admissionAPIVersion,
			Kind:       admissionKind,
		},
	}

	// Serialize the admission response
	resp, err := encodeAdmissionReview(responseReview)
	if err != nil {
		klog.Errorf("Error encoding admission response: %v", err)
		http.Error(w, "Error encoding admission response", http.StatusInternalServerError)
		return
	}

	// Write the admission response
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Error writing admission response: %v", err)
		http.Error(w, "Error writing admission response", http.StatusInternalServerError)
		return
	}

	klog.Infof("Validation Status: %v", response.Allowed)
}

func decodeAdmissionReview(body []byte) (v1.AdmissionReview, error) {
	admissionReview := v1.AdmissionReview{}
	codecs := serializer.NewCodecFactory(runtime.NewScheme())
	if _, _, err := codecs.UniversalDeserializer().Decode(body, nil, &admissionReview); err != nil {
		return admissionReview, err
	}
	return admissionReview, nil
}

func encodeAdmissionReview(responseReview v1.AdmissionReview) ([]byte, error) {
	resp, err := json.Marshal(responseReview)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func handleAdmissionRequest(admissionReview v1.AdmissionReview, _ KubernetesClient, isTodayFriday func() bool) *v1.AdmissionResponse {
	var deploymentName, namespace string
	if admissionReview.Request != nil && admissionReview.Request.Object.Raw != nil {
		deployment := &appsv1.Deployment{}
		if err := json.Unmarshal(admissionReview.Request.Object.Raw, deployment); err != nil {
			klog.Errorf("Error decoding deployment: %v", err)
			return makeAdmissionResponse(admissionReview.Request.UID, false, "Error decoding deployment object.")
		}
		deploymentName = deployment.Name
		namespace = deployment.Namespace
	}

	if isTodayFriday() {
		klog.Infof("Denying the request to create a new Deployment on Friday. Deployment: %s, Namespace: %s", deploymentName, namespace)
		return makeAdmissionResponse(admissionReview.Request.UID, false, "Creating new Deployments on Fridays is not allowed.")
	}

	klog.Infof("Allowing the request to create a new Deployment. Deployment: %s, Namespace: %s", deploymentName, namespace)

	return makeAdmissionResponse(admissionReview.Request.UID, true, "")
}

func makeAdmissionResponse(uid types.UID, allowed bool, message string) *v1.AdmissionResponse {
	return &v1.AdmissionResponse{
		UID:     uid,
		Allowed: allowed,
		Result: &metav1.Status{
			Message: message,
		},
	}
}

func isFriday() bool {
	return time.Now().Weekday() == time.Friday
}
