package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/klog"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/tools/clientcmd"

	v1 "github.com/petrkotas/k8s-object-lock/pkg/api/lock/v1"
	clientset "github.com/petrkotas/k8s-object-lock/pkg/generated/clientset/versioned"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()
)

// Conf caries client configuration
type Conf struct {
	Client *clientset.Clientset
}

// MakeServerConf Creates the client configuration
func MakeServerConf(k8sMasterURL, kubeconfig string) *Conf {
	cfg, err := clientcmd.BuildConfigFromFlags(k8sMasterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	lockClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building example clientset: %s", err.Error())
	}

	return &Conf{
		Client: lockClient,
	}
}

func itemNotInList(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return false
		}
	}

	return true
}

// validate takes the object on admission request and checks whether the same
// object is annotated in the etcd as locked.
// If objects exist in etcd and is locked => request fails
// otherwise request passes
func checkLock(lock *v1.Lock, admissionReview *v1beta1.AdmissionReview) bool {
	klog.Info("Processing the request")

	// only when lock is returned it marks the object for lockdown
	if lock != nil {
		klog.Infof("Lock received: %s/%s", lock.Namespace, lock.Name)

		// check the rules whether the object is indeed locked on the operation

		locked := true

		if lock.Spec.Operations != nil {
			if itemNotInList(string(admissionReview.Request.Operation), lock.Spec.Operations) {
				klog.Infof("Operation %s not found in locked operations", string(admissionReview.Request.Operation))
				locked = false
			}
		}

		if lock.Spec.APIGroups != nil {
			if itemNotInList(admissionReview.Request.Resource.Group, lock.Spec.APIGroups) {
				klog.Infof("APIGroup %s not found in locked operations", string(admissionReview.Request.Resource.Group))
				locked = false
			}
		}

		if lock.Spec.APIVersions != nil {
			if itemNotInList(admissionReview.Request.Resource.Version, lock.Spec.APIVersions) {
				klog.Infof("APIVersion %s not found in locked operations", string(admissionReview.Request.Resource.Version))
				locked = false
			}
		}

		if lock.Spec.Resources != nil {
			if itemNotInList(admissionReview.Request.Resource.Resource, lock.Spec.Resources) {
				klog.Infof("Resource %s not found in locked operations", string(admissionReview.Request.Resource.Resource))
				locked = false
			}
		}

		if lock.Spec.SubResources != nil {
			if itemNotInList(admissionReview.Request.SubResource, lock.Spec.SubResources) {
				klog.Infof("SubResource %s not found in locked operations", string(admissionReview.Request.SubResource))
				locked = false
			}
		}

		if locked {
			return true
		}
	}

	return false
}

// Validate process the request, parse the data from it and pass to check
func (s *Conf) Validate(w http.ResponseWriter, r *http.Request) {
	klog.Info("Validating request")
	ctx := r.Context()

	// parse incoming request => admission request
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		klog.Error("Empty body")
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		klog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		var result *metav1.Status
		allowed := true
		admissionResponse = &v1beta1.AdmissionResponse{
			Allowed: allowed,
			Result:  result,
		}

		// check if there is lock object in the same namespace with the same name
		// If so, than fail
		kind := ar.Request.Kind.String()
		name := ar.Request.Name
		namespace := ar.Request.Namespace

		// directly ask the API. The calls should be so sparse, there is no reason in using cached listers.
		klog.Infof("Looking for a lock: %s - %s/%s", kind, namespace, name)
		klog.Infof("Admission request: %s, %s, %s", ar.Request.Resource, ar.Request.SubResource, ar.Request.Operation)

		var lock *v1.Lock
		lock, err = s.Client.LocksV1().Locks(ar.Request.Namespace).Get(ctx, ar.Request.Name, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			klog.Infof("Lock: %s/%s not found", ar.Request.Namespace, ar.Request.Name)
			lock = nil
		} else if err != nil {
			klog.Errorf("An error occured, the admission review will not be done: %s", err.Error())
			lock = nil
		}

		if checkLock(lock, &ar) {
			admissionResponse.Allowed = false
			admissionResponse.Result = &metav1.Status{
				Message: fmt.Sprintf("Object %s/%s is locked, reason: %s", lock.Namespace, lock.Name, lock.Spec.Reason),
			}
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		klog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	klog.Infof("Ready to write reponse ...")

	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
	w.Header().Add("Content-Type", "application/json")
	if _, err := w.Write(resp); err != nil {
		klog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}
