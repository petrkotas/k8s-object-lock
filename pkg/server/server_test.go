package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/petrkotas/k8s-object-lock/pkg/api/lock/v1"

	admv1 "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	k8sv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Lock check", func() {

	var lock v1.Lock
	var admissionReview admv1.AdmissionReview

	// deployment is the object used for testing the locking
	deployment := appsv1.Deployment{
		ObjectMeta: k8sv1.ObjectMeta{
			Name:      "locked-object",
			Namespace: "default",
		},
	}

	admissionReview = admv1.AdmissionReview{
		Request: &admv1.AdmissionRequest{
			Name:      "locked-object",
			Namespace: "default",
			Kind: k8sv1.GroupVersionKind{
				Group:   deployment.GroupVersionKind().Group,
				Version: deployment.GroupVersionKind().Version,
				Kind:    deployment.GroupVersionKind().Kind,
			},
			Operation:   admv1.Delete,
			SubResource: "scale",
		},
	}

	Context("with valid lock", func() {
		BeforeEach(func() {
			// lock is the base lock that is being used for the whole tests
			lock = v1.Lock{
				ObjectMeta: k8sv1.ObjectMeta{
					Name:      "locked-object",
					Namespace: "default",
				},
			}
		})

		It("should verify the object as locked", func() {
			Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
		})

		Context("with selected operation", func() {
			It("should be locked when operation match", func() {
				lock.Spec = v1.LockSpec{
					Operations: []string{"DELETE"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
			})

			It("should be unlocked when operation do not match", func() {
				lock.Spec = v1.LockSpec{
					Operations: []string{"UPDATE"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})

		})

		Context("with selected api version", func() {
			It("should be locked when api version match", func() {
				lock.Spec = v1.LockSpec{
					APIVersions: []string{deployment.GroupVersionKind().Version},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
			})

			It("should be unlocked when operation do not match", func() {
				lock.Spec = v1.LockSpec{
					APIVersions: []string{"v12gamma3"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})
		})

		Context("with selected api group", func() {
			It("should be locked when api group match", func() {
				lock.Spec = v1.LockSpec{
					APIGroups: []string{deployment.GroupVersionKind().Group},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
			})

			It("should be unlocked when api group do not match", func() {
				lock.Spec = v1.LockSpec{
					APIGroups: []string{"greatGroup"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})
		})

		Context("with selected resource", func() {
			It("should be locked when resource match", func() {
				lock.Spec = v1.LockSpec{
					Resources: []string{deployment.GroupVersionKind().Kind},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
			})

			It("should be unlocked when resource do not match", func() {
				lock.Spec = v1.LockSpec{
					Resources: []string{"pod"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})
		})

		Context("with selected subresource", func() {
			It("should be locked when subresource match", func() {
				lock.Spec = v1.LockSpec{
					SubResources: []string{admissionReview.Request.SubResource},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
			})

			It("should be unlocked when subresource do not match", func() {
				lock.Spec = v1.LockSpec{
					Resources: []string{"logs"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})
		})

		Context("with all properties set", func() {
			It("should be locked when all properties match", func() {
				lock.Spec = v1.LockSpec{
					Operations:   []string{"DELETE"},
					APIVersions:  []string{deployment.GroupVersionKind().Version},
					APIGroups:    []string{deployment.GroupVersionKind().Group},
					Resources:    []string{deployment.GroupVersionKind().Kind},
					SubResources: []string{admissionReview.Request.SubResource},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeTrue(), "the object should be considered locked")
			})

			It("should be unlocked when even one does not match", func() {
				lock.Spec = v1.LockSpec{
					Operations:   []string{"DELETE"},
					APIVersions:  []string{deployment.GroupVersionKind().Version},
					APIGroups:    []string{deployment.GroupVersionKind().Group},
					Resources:    []string{deployment.GroupVersionKind().Kind},
					SubResources: []string{"logs"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})

			It("should be unlocked when all not match", func() {
				lock.Spec = v1.LockSpec{
					Operations:   []string{"UPDATE"},
					APIVersions:  []string{"v12gamma3"},
					APIGroups:    []string{"greatGroup"},
					Resources:    []string{"pod"},
					SubResources: []string{"logs"},
				}

				Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
			})

		})

	})

	Context("with nil lock", func() {
		It("should verify the object as not locked", func() {
			Expect(checkLock(&lock, &admissionReview)).To(BeFalse(), "the object should be considered unlocked")
		})
	})

})
