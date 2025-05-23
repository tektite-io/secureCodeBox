// SPDX-FileCopyrightText: the secureCodeBox authors
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-logr/logr"

	config "github.com/secureCodeBox/secureCodeBox/auto-discovery/kubernetes/pkg/config"
	"github.com/secureCodeBox/secureCodeBox/auto-discovery/kubernetes/pkg/util"
	executionv1 "github.com/secureCodeBox/secureCodeBox/operator/apis/execution/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ContainerScanReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Config   config.AutoDiscoveryConfig
}

type Cluster struct {
	Name string
}
type ContainerAutoDiscoveryTemplateArgs struct {
	Config     config.AutoDiscoveryConfig
	ScanConfig config.ScanConfig
	Cluster    config.ClusterConfig
	Target     metav1.ObjectMeta
	Namespace  corev1.Namespace
	ImageID    string
}

// +kubebuilder:rbac:groups="execution.securecodebox.io",resources=scantypes,verbs=get;list;watch
// +kubebuilder:rbac:groups="execution.securecodebox.io",resources=scheduledscans,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="execution.securecodebox.io/status",resources=scheduledscans,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile compares the Pod object against the state of the cluster and updates both if needed
func (r *ContainerScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log

	log.V(8).Info("Something happened to a pod", "pod", req.Name, "namespace", req.Namespace)

	var pod corev1.Pod
	err := r.Get(ctx, req.NamespacedName, &pod)
	if err != nil {
		//dont log an error, as a deleted pod cant be fetched and would spam the logs
		log.V(7).Info("Unable to fetch Pod", "pod", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if podManagedBySecureCodeBox(pod) {
		log.V(7).Info("Pod will be ignored, as it is managed by securecodebox", "pod", pod)
		return ctrl.Result{}, nil
	}

	if podNotReady(pod) {
		log.V(6).Info("Pod not ready", "pod", pod.Name)
		return ctrl.Result{}, nil
	}

	scanTypesInstalled := r.checkForScanTypes(ctx, pod)
	if !scanTypesInstalled {
		requeueDuration := r.Config.ContainerAutoDiscovery.PassiveReconcileInterval.Duration
		return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
	}

	if pod.DeletionTimestamp == nil {
		r.checkIfNewScansNeedToBeCreated(ctx, pod)
	} else {
		r.checkIfScansNeedToBeDeleted(ctx, pod)
	}

	return ctrl.Result{}, nil
}

func podManagedBySecureCodeBox(pod corev1.Pod) bool {
	labels := pod.ObjectMeta.Labels
	value, exists := labels["app.kubernetes.io/managed-by"]

	if exists {
		regexPattern := regexp.MustCompile("securecodebox.*")
		return regexPattern.MatchString(value)
	}
	return false
}

func podNotReady(pod corev1.Pod) bool {
	//check if container imageIDs are present, otherwise pod is not ready yet
	return len(getImageIDsForPod(pod)) == 0
}

func (r *ContainerScanReconciler) checkIfNewScansNeedToBeCreated(ctx context.Context, pod corev1.Pod) {
	r.Log.V(8).Info("Pod is running", "pod", pod.Name, "namespace", pod.Namespace)
	nonScannedImageIDs := r.getNonScannedImageIDs(ctx, pod)
	//log if there are any unscanned containers
	if len(nonScannedImageIDs) > 0 {
		r.Log.Info("Discovered one or more new unscanned containers; scanning them now", "pod", pod.Name, "namespace", pod.Namespace)
	}
	r.createScheduledScans(ctx, pod, nonScannedImageIDs)
}

func (r *ContainerScanReconciler) getNonScannedImageIDs(ctx context.Context, pod corev1.Pod) map[string][]config.ScanConfig {
	result := make(map[string][]config.ScanConfig)
	allImageIDs := getImageIDsForPod(pod)

	for _, imageID := range allImageIDs {
		cleanedImageID := cleanupImageID(imageID)

		if len(r.Config.ContainerAutoDiscovery.ScanConfigs) == 0 {
			r.Log.Info("Warning: You have the Container AutoDiscovery enabled but don't have any `scanConfigs` in your AutoDiscovery configuration. No scans will be started!")
		}
		for _, scanConfig := range r.Config.ContainerAutoDiscovery.ScanConfigs {
			scanName := getScanName(cleanedImageID, scanConfig)

			var scan executionv1.ScheduledScan
			err := r.Client.Get(ctx, types.NamespacedName{Name: scanName, Namespace: pod.Namespace}, &scan)

			if k8sErrors.IsNotFound(err) {
				result[cleanedImageID] = append(result[cleanedImageID], scanConfig)
			} else if err != nil {
				r.Log.Error(err, "Unable to fetch scan", "name", scanName, "namespace", pod.Namespace)
			}
		}
	}
	return result
}

func getImageIDsForPod(pod corev1.Pod) []string {
	var result []string

	for _, container := range pod.Status.ContainerStatuses {
		imageID := container.ImageID
		if imageID != "" {
			result = append(result, imageID)
		}
	}
	return result
}

func cleanupImageID(imageID string) string {
	//some setups will add a protocol to the imageID like "docker-pullable://some/image:sha256@0123456789"
	//this function removes the protocol as it interferes with trivy
	//the imageid above will be transformed to "some/image:sha256@0123456789"
	imageRegex := regexp.MustCompile("(.*://)?(.*)")
	return imageRegex.FindStringSubmatch(imageID)[2]
}

func getScanName(imageID string, scanConfig config.ScanConfig) string {
	//function builds string like: _appName_-_customScanName_-at-_imageID_HASH_ eg: nginx-myTrivyScan-at-0123456789

	hashRegex := regexp.MustCompile(".*/(?P<appName>.*)@sha256:(?P<hash>.*)")
	appName := hashRegex.FindStringSubmatch(imageID)[1]
	hash := hashRegex.FindStringSubmatch(imageID)[2]

	//cutoff appname if it is longer than 20 chars
	maxAppLength := 20
	if len(appName) > maxAppLength {
		appName = appName[:maxAppLength]
	}

	result := fmt.Sprintf("%s-%s-at-%s", appName, scanConfig.Name, hash)

	result = strings.ReplaceAll(result, ".", "-")
	result = strings.ReplaceAll(result, "/", "-")
	result = strings.ReplaceAll(result, "_", "-")

	//limit scan name length to kubernetes limits
	return result[:62]
}

func (r *ContainerScanReconciler) createScheduledScans(ctx context.Context, pod corev1.Pod, scansToBeCreated map[string][]config.ScanConfig) {
	secretsDefined, secrets := checkForImagePullSecrets(pod)
	mapSecretsToEnv := r.Config.ContainerAutoDiscovery.ImagePullSecretConfig.MapImagePullSecretsToEnvironmentVariables

	for imageID, scans := range scansToBeCreated {
		for _, scan := range scans {
			if secretsDefined {
				if mapSecretsToEnv {
					r.createScheduledScanWithImagePullSecrets(ctx, pod, imageID, scan, secrets)
				}
			} else {
				r.createScheduledScan(ctx, pod, imageID, scan)
			}
		}
	}
}

func checkForImagePullSecrets(pod corev1.Pod) (bool, []corev1.LocalObjectReference) {
	imagePullSecrets := pod.Spec.ImagePullSecrets
	secretsDefined := len(imagePullSecrets) > 0
	return secretsDefined, imagePullSecrets
}

func (r *ContainerScanReconciler) createScheduledScan(ctx context.Context, pod corev1.Pod, imageID string, scanConfig config.ScanConfig) {
	namespace := r.getNamespace(ctx, pod)

	newScheduledScan := r.generateScan(pod, imageID, scanConfig, namespace)
	err := r.Client.Create(ctx, &newScheduledScan)

	if err != nil {
		r.Log.Error(err, "Failed to create scheduled scan", "scan", newScheduledScan, "namespace", namespace)
	} else {
		r.Log.V(6).Info("Created scheduled scan", "pod", pod.Name, "namespace", pod.Namespace)
	}
}

func (r *ContainerScanReconciler) createScheduledScanWithImagePullSecrets(ctx context.Context, pod corev1.Pod, imageID string, scanConfig config.ScanConfig, secrets []corev1.LocalObjectReference) {
	namespace := r.getNamespace(ctx, pod)

	newScheduledScan := r.generateScanWithVolumeMounts(pod, imageID, scanConfig, namespace, secrets)
	err := r.Client.Create(ctx, &newScheduledScan)

	if err != nil {
		r.Log.Error(err, "Failed to create scheduled scan", "scan", newScheduledScan, "namespace", namespace)
	} else {
		r.Log.V(6).Info("Created scheduled scan", "pod", pod.Name, "namespace", pod.Namespace)
	}
}

func (r *ContainerScanReconciler) getNamespace(ctx context.Context, pod corev1.Pod) corev1.Namespace {
	var result corev1.Namespace
	err := r.Client.Get(ctx, types.NamespacedName{Name: pod.Namespace, Namespace: ""}, &result)

	if err != nil {
		r.Log.Error(err, "Could not retrieve namespace of pod", "pod", pod)
	}
	return result
}

func (r *ContainerScanReconciler) generateScan(pod corev1.Pod, imageID string, scanConfig config.ScanConfig, namespace corev1.Namespace) executionv1.ScheduledScan {
	templateArgs := ContainerAutoDiscoveryTemplateArgs{
		Config:     r.Config,
		ScanConfig: scanConfig,
		Cluster:    r.Config.Cluster,
		Target:     pod.ObjectMeta,
		Namespace:  namespace,
		ImageID:    imageID,
	}
	scanSpec := util.GenerateScanSpec(scanConfig, templateArgs)

	newScheduledScan := executionv1.ScheduledScan{
		ObjectMeta: metav1.ObjectMeta{
			Name:        getScanName(imageID, scanConfig),
			Namespace:   pod.Namespace,
			Annotations: getScanAnnotations(scanConfig, templateArgs),
			Labels:      getScanLabels(scanConfig, templateArgs),
		},
		Spec: scanSpec,
	}
	return newScheduledScan
}

func (r *ContainerScanReconciler) generateScanWithVolumeMounts(pod corev1.Pod, imageID string, scanConfig config.ScanConfig, namespace corev1.Namespace, secrets []corev1.LocalObjectReference) executionv1.ScheduledScan {
	templateArgs := ContainerAutoDiscoveryTemplateArgs{
		Config:     r.Config,
		ScanConfig: scanConfig,
		Cluster:    r.Config.Cluster,
		Target:     pod.ObjectMeta,
		Namespace:  namespace,
		ImageID:    imageID,
	}

	extraVolumes, extraMounts := getVolumesForSecrets(secrets)
	scanConfig.Volumes = append(scanConfig.Volumes, extraVolumes...)

	scanSpec := util.GenerateScanSpec(scanConfig, templateArgs)
	scanSpec.ScanSpec.InitContainers = append(scanSpec.ScanSpec.InitContainers, getSecretExtractionInitContainer(imageID, scanConfig, extraMounts))

	pullSecretConfig := r.Config.ContainerAutoDiscovery.ImagePullSecretConfig
	usernameEnvVarName := pullSecretConfig.UsernameEnvironmentVariableName
	passwordEnvVarName := pullSecretConfig.PasswordNameEnvironmentVariableName
	scanSpec.ScanSpec.Env = append(scanSpec.ScanSpec.Env, getTemporarySecretEnvironmentVariableMount(imageID, scanConfig, usernameEnvVarName, passwordEnvVarName)...)

	newScheduledScan := executionv1.ScheduledScan{
		ObjectMeta: metav1.ObjectMeta{
			Name:        getScanName(imageID, scanConfig),
			Namespace:   pod.Namespace,
			Annotations: getScanAnnotations(scanConfig, templateArgs),
			Labels:      getScanLabels(scanConfig, templateArgs),
		},
		Spec: scanSpec,
	}
	return newScheduledScan
}

func getVolumesForSecrets(secrets []corev1.LocalObjectReference) ([]corev1.Volume, []corev1.VolumeMount) {
	var volumes []corev1.Volume
	var mounts []corev1.VolumeMount
	for _, secret := range secrets {
		volume := corev1.Volume{
			Name: secret.Name + "-volume",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secret.Name,
				},
			},
		}

		mount := corev1.VolumeMount{
			Name:      secret.Name + "-volume",
			MountPath: "/secrets/" + secret.Name,
		}

		volumes = append(volumes, volume)
		mounts = append(mounts, mount)
	}
	return volumes, mounts
}

func getSecretExtractionInitContainer(imageID string, scanConfig config.ScanConfig, volumeMounts []corev1.VolumeMount) corev1.Container {
	temporarySecretName := getTemporarySecretName(imageID, scanConfig)

	return corev1.Container{
		Name:         "secret-extraction-to-env",
		Image:        "docker.io/securecodebox/auto-discovery-pull-secret-extractor",
		Command:      []string{"python"},
		Args:         []string{"secret_extraction.py", imageID, temporarySecretName},
		VolumeMounts: volumeMounts,
		Env: []corev1.EnvVar{
			{
				Name: "POD_NAME",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.name",
					},
				},
			},
			{
				Name: "NAMESPACE",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "metadata.namespace",
					},
				},
			},
		},
	}
}

func getTemporarySecretName(imageID string, scanConfig config.ScanConfig) string {
	//limit name to kubernetes max length
	return ("temporary-secret-" + getScanName(imageID, scanConfig))[:62]
}

func getTemporarySecretEnvironmentVariableMount(imageID string, scanConfig config.ScanConfig, usernameEnvVarName string, passwordEnvVarName string) []corev1.EnvVar {
	trueBool := true
	return []corev1.EnvVar{
		{
			Name: usernameEnvVarName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Optional: &trueBool,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: getTemporarySecretName(imageID, scanConfig),
					},
					Key: "username",
				},
			},
		},
		{
			Name: passwordEnvVarName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Optional: &trueBool,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: getTemporarySecretName(imageID, scanConfig),
					},
					Key: "password",
				},
			},
		},
	}
}

func (r *ContainerScanReconciler) checkForScanTypes(ctx context.Context, pod corev1.Pod) bool {
	result := true
	namespace := r.getNamespace(ctx, pod)

	scanConfigs := r.Config.ContainerAutoDiscovery.ScanConfigs
	for _, scanConfig := range scanConfigs {
		scanTypeName := scanConfig.ScanType
		scanType := executionv1.ScanType{}
		err := r.Get(ctx, types.NamespacedName{Name: scanTypeName, Namespace: namespace.Name}, &scanType)
		if k8sErrors.IsNotFound(err) {
			r.Log.Info("Namespace requires configured ScanType to properly start automatic scans.", "namespace", namespace.Name, "service", pod.Name, "scanType", scanTypeName)
			// Add event to pod to communicate failure to user
			r.Recorder.Event(&pod, "Warning", "ScanTypeMissing", "Namespace requires ScanType '"+scanTypeName+"' to properly start automatic scans.")
			result = false
		}
	}
	return result
}

func (r *ContainerScanReconciler) checkIfScansNeedToBeDeleted(ctx context.Context, pod corev1.Pod) {
	r.Log.V(8).Info("Pod will be deleted", "pod", pod.Name, "namespace", pod.Namespace, "timestamp", pod.DeletionTimestamp)
	allImageIDs := getImageIDsForPod(pod)
	imageIDsToBeDeleted := r.getOrphanedScanImageIDs(ctx, pod, allImageIDs)
	//log if there are any orphaned scans
	if len(imageIDsToBeDeleted) > 0 {
		r.Log.Info("Discovered one or more 'Trivy' scans related to a non-active container; deleting them now", "pod", pod.Name, "namespace", pod.Namespace)
	}
	r.deleteScans(ctx, pod, imageIDsToBeDeleted)
}

func (r *ContainerScanReconciler) getOrphanedScanImageIDs(ctx context.Context, pod corev1.Pod, imageIDs []string) map[string][]config.ScanConfig {
	result := make(map[string][]config.ScanConfig)

	for _, imageID := range imageIDs {
		cleanedImageID := cleanupImageID(imageID)

		for _, scanConfig := range r.Config.ContainerAutoDiscovery.ScanConfigs {
			scanName := getScanName(cleanedImageID, scanConfig)

			var scan executionv1.ScheduledScan
			err := r.Client.Get(ctx, types.NamespacedName{Name: scanName, Namespace: pod.Namespace}, &scan)
			if err != nil {
				if !k8sErrors.IsNotFound(err) {
					r.Log.Error(err, "Unable to fetch scan", "name", scanName)
				}
			} else if !r.containerIDInUse(ctx, pod, imageID) {
				result[cleanedImageID] = append(result[cleanedImageID], scanConfig)
			}
		}
	}
	return result
}

func (r *ContainerScanReconciler) containerIDInUse(ctx context.Context, pod corev1.Pod, imageID string) bool {
	var pods corev1.PodList
	err := r.Client.List(ctx, &pods, client.InNamespace(pod.Namespace))

	if err != nil {
		r.Log.Error(err, "Unable to fetch pods from namespace", "namespace", pod.Namespace)
		return false
	}
	return searchForImageIDInPods(pods.Items, imageID)
}

func searchForImageIDInPods(pods []corev1.Pod, targetImageID string) bool {
	for _, pod := range pods {
		imageIDS := getImageIDsForPod(pod)

		for _, imageID := range imageIDS {
			if pod.DeletionTimestamp == nil && imageID == targetImageID {
				return true
			}
		}
	}
	return false
}

func (r *ContainerScanReconciler) deleteScans(ctx context.Context, pod corev1.Pod, scansToBeDeleted map[string][]config.ScanConfig) {
	for imageID, scanConfigs := range scansToBeDeleted {
		for _, scanConfig := range scanConfigs {

			scan, err := r.getScan(ctx, pod, imageID, scanConfig)

			if err != nil {
				r.Log.Error(err, "Unable to fetch scan", "pod", pod, "imageID", imageID)
				continue
			}

			err = r.Client.Delete(ctx, &scan)
			if err != nil {
				r.Log.Error(err, "Unable to delete scheduled scan", "scan", scan.Name)
			} else {
				r.Log.V(6).Info("Deleting scheduled scan", "scan", scan.Name)
			}
		}
	}
}

func (r *ContainerScanReconciler) getScan(ctx context.Context, pod corev1.Pod, imageID string, scanConfig config.ScanConfig) (executionv1.ScheduledScan, error) {
	var scan executionv1.ScheduledScan
	scanName := getScanName(imageID, scanConfig)
	err := r.Client.Get(ctx, types.NamespacedName{Name: scanName, Namespace: pod.Namespace}, &scan)
	return scan, err
}

func getScanAnnotations(scanConfig config.ScanConfig, templateArgs ContainerAutoDiscoveryTemplateArgs) map[string]string {
	return util.ParseMapTemplate(templateArgs, scanConfig.Annotations)
}

func getScanLabels(scanConfig config.ScanConfig, templateArgs ContainerAutoDiscoveryTemplateArgs) map[string]string {
	generatedLabels := util.ParseMapTemplate(templateArgs, scanConfig.Labels)
	generatedLabels["app.kubernetes.io/managed-by"] = "securecodebox-autodiscovery"

	return generatedLabels
}

// SetupWithManager sets up the controller and initializes every thing it needs
func (r *ContainerScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	err := util.CheckUniquenessOfScanNames(r.Config.ContainerAutoDiscovery.ScanConfigs)
	if err != nil {
		r.Log.Error(err, "Scan names are not unique")
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(util.GetPredicates(mgr.GetClient(), r.Log, r.Config.ResourceInclusion.Mode)).
		Complete(r)
}
