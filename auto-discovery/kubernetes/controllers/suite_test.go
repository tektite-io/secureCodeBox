// SPDX-FileCopyrightText: the secureCodeBox authors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	ctrl "sigs.k8s.io/controller-runtime"

	//+kubebuilder:scaffold:imports

	executionv1 "github.com/secureCodeBox/secureCodeBox/operator/apis/execution/v1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "operator", "crds")},
		ErrorIfCRDPathMissing: false,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = executionv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	// working config
	err = (&ServiceScanReconciler{
		Client:   k8sManager.GetClient(),
		Scheme:   k8sManager.GetScheme(),
		Recorder: k8sManager.GetEventRecorderFor("ServiceScanController"),
		Log:      ctrl.Log.WithName("controllers").WithName("ServiceScanController"),
		Config:   AutoDiscoveryConfigMock,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	// working config
	err = (&ContainerScanReconciler{
		Client:   k8sManager.GetClient(),
		Scheme:   k8sManager.GetScheme(),
		Recorder: k8sManager.GetEventRecorderFor("ContainerScanController"),
		Log:      ctrl.Log.WithName("controllers").WithName("ContainerScanController"),
		Config:   AutoDiscoveryConfigMock,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()
}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
