//+build e2e

package tests

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	lockv1 "github.com/petrkotas/k8s-object-lock/pkg/api/lock/v1"
)

var _ = Describe("with kubernetes lock", func() {

	parseFlags()

	kubeCli, lockCli := createClientSets(kubeconfig)

	var replicas int32 = 1
	nginxDeploy := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nginx",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:  "naginx",
							Image: "nginx:1.7.9",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	lock := lockv1.Lock{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: "default",
		},
	}

	dplClient := kubeCli.AppsV1().Deployments("default")
	lockClient := lockCli.LocksV1().Locks("default")

	ctx := context.Background()

	BeforeEach(func() {
		Eventually(func() int {
			pods, err := kubeCli.CoreV1().Pods("default").List(ctx, metav1.ListOptions{FieldSelector: "status.phase=Running"})
			Expect(err).ToNot(HaveOccurred(), "should list pods in default namespace")
			return len(pods.Items)
		}, 60*time.Second).Should(Equal(0), "no pods should be running after tests")
	})

	AfterEach(func() {
		// cleanup namespace
		err := lockCli.LocksV1().RESTClient().Delete().Namespace("default").Resource("locks").Do(ctx).Error()
		Expect(err).ToNot(HaveOccurred(), "should list locks without problem")
		Eventually(func() int {
			locks, err := lockClient.List(ctx, metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred(), "should list locks in default namespace")
			return len(locks.Items)
		}, 60*time.Second).Should(Equal(0), "no locks should be present before tests")

		err = kubeCli.AppsV1().RESTClient().Delete().Namespace("default").Resource("deployments").Do(ctx).Error()
		Expect(err).ToNot(HaveOccurred(), "should list deployments without problem")
		Eventually(func() int {
			dpls, err := dplClient.List(ctx, metav1.ListOptions{})
			Expect(err).ToNot(HaveOccurred(), "should list deployments in default namespace")
			return len(dpls.Items)
		}, 60*time.Second).Should(Equal(0), "no deployments should be running before tests")
	})

	Context("with lock deployed", func() {
		When("namespace is labeled", func() {

			BeforeEach(func() {
				namespace, err := kubeCli.CoreV1().Namespaces().Patch(ctx, "default", types.MergePatchType, []byte(
					`{"metadata":{"labels":{"lockable":"true"}}}`), metav1.PatchOptions{})
				Expect(err).ToNot(HaveOccurred(), "should patch namespace")
				Expect(namespace.GetObjectMeta().GetLabels()).To(Equal(map[string]string{"kubernetes.io/metadata.name": "default", "lockable": "true"}), "no labels should be on the namespace")
			})

			It("should block creation", func() {
				// create lock
				newLock, err := lockClient.Create(ctx, &lock, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "Lock should be created without any issue")
				Expect(newLock.Name).To(Equal(lock.Name), "Lock should have the same name")

				// try to create deployment
				_, err = dplClient.Create(ctx, &nginxDeploy, metav1.CreateOptions{})
				Expect(err).To(HaveOccurred(), "lock should block creation")
			})

			It("should block deletion", func() {
				// create deployment
				newDeploy, err := dplClient.Create(ctx, &nginxDeploy, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "should create ok")

				// create lock
				newLock, err := lockClient.Create(ctx, &lock, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "Lock should be created without any issue")
				Expect(newLock.Name).To(Equal(lock.Name), "Lock should have the same name")

				// delete deployment
				err = dplClient.Delete(ctx, newDeploy.Name, metav1.DeleteOptions{})
				Expect(err).To(HaveOccurred(), "lock should block deletion")
			})

			It("should block scale subresource", func() {
				// create deployment
				_, err := dplClient.Create(ctx, &nginxDeploy, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "should create ok")

				// create lock
				newLock, err := lockClient.Create(ctx, &lock, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "Lock should be created without any issue")
				Expect(newLock.Name).To(Equal(lock.Name), "Lock should have the same name")

				scale, err := dplClient.GetScale(ctx, nginxDeploy.Name, metav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred(), "Should get scale")

				scale.Spec.Replicas = 3
				_, err = dplClient.UpdateScale(ctx, nginxDeploy.Name, scale, metav1.UpdateOptions{})
				Expect(err).To(HaveOccurred(), "scale subresource update should be blocked")
			})
		})

		When("namespace is not labeled", func() {

			BeforeEach(func() {
				namespace, err := kubeCli.CoreV1().Namespaces().Patch(ctx, "default", types.MergePatchType, []byte(
					`{"metadata":{"labels":{"lockable":"false"}}}`), metav1.PatchOptions{})
				Expect(err).ToNot(HaveOccurred(), "should patch namespace")
				Expect(namespace.GetObjectMeta().GetLabels()).To(Equal(map[string]string{"kubernetes.io/metadata.name": "default", "lockable": "false"}), "no labels should be on the namespace")
			})

			It("should not affect deletion", func() {
				// create lock
				newLock, err := lockClient.Create(ctx, &lock, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "Lock should be created without any issue")
				Expect(newLock.Name).To(Equal(lock.Name), "Lock should have the same name")

				// create deployment
				_, err = dplClient.Create(ctx, &nginxDeploy, metav1.CreateOptions{})
				Expect(err).ToNot(HaveOccurred(), "should create ok")
			})

		})
	})

	Context("without lock deployed", func() {
		It("should not block creation", func() {
			// create deployment
			_, err := dplClient.Create(ctx, &nginxDeploy, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred(), "lock should block creation")
		})

		It("should not block scale subresource", func() {
			// create deployment
			_, err := dplClient.Create(ctx, &nginxDeploy, metav1.CreateOptions{})
			Expect(err).ToNot(HaveOccurred(), "should create ok")

			Eventually(func() int {
				pods, err := kubeCli.CoreV1().Pods("default").List(ctx, metav1.ListOptions{FieldSelector: "status.phase=Running"})
				Expect(err).ToNot(HaveOccurred(), "should list pods without problems")
				return len(pods.Items)
			}, 60*time.Second).Should(Equal(1), "deployment should be 1")

			scale, err := dplClient.GetScale(ctx, nginxDeploy.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred(), "Should get scale")

			scale.Spec.Replicas = 3
			_, err = dplClient.UpdateScale(ctx, nginxDeploy.Name, scale, metav1.UpdateOptions{})
			Expect(err).ToNot(HaveOccurred(), "scale subresource update should not be blocked")

			Eventually(func() int {
				pods, err := kubeCli.CoreV1().Pods("default").List(ctx, metav1.ListOptions{FieldSelector: "status.phase=Running"})
				Expect(err).ToNot(HaveOccurred(), "should list pods without problems")
				return len(pods.Items)
			}, 60*time.Second).Should(Equal(3), "deployment shoudl scale to 3")
		})
	})
})
