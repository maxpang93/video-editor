package ffmpegworker

import (
	"context"
	"fmt"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var K8sClient *kubernetes.Clientset

const (
	LiteJobCPURequest  = "1000m"
	LiteJobMemRequest  = "1Gi"
	HeavyJobCPURequest = "2000m"
	HeavyJobMemRequest = "2Gi"
)

func init() {
	var err error
	K8sClient, err = NewK8sClient()
	if err != nil {
		panic(err)
	}

	// BuildWorkerImage(context.Background(), "Dockerfile.worker")
}

func NewK8sClient() (*kubernetes.Clientset, error) {
	kubeconfig := os.Getenv("K3S_KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig: %w", err)
	}

	config.Host = fmt.Sprintf("https://%s:6443", os.Getenv("K3S_MASTER_IP"))

	return kubernetes.NewForConfig(config)
}

func newJob(command []string, cpuRequest, memRequest string) *batchv1.Job {
	backoffLimit := int32(1)
	TTLSecondsAfterFinished := int32(60)
	hostPathType := corev1.HostPathDirectory
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "ffmpeg-job-",
			Namespace:    "default",
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &backoffLimit,
			TTLSecondsAfterFinished: &TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{
						{
							Name: "media",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/media",
									Type: &hostPathType,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "ffmpeg",
							Image:           "ffmpeg-worker:latest",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command:         command,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "media",
									MountPath: "/video",
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(cpuRequest),
									corev1.ResourceMemory: resource.MustParse(memRequest),
								},
							},
						},
					},
				},
			},
		},
	}
	return job
}

func CreateLiteJob(command []string) error {
	job := newJob(command, LiteJobCPURequest, LiteJobMemRequest)
	_, err := K8sClient.BatchV1().Jobs("default").Create(
		context.Background(),
		job,
		metav1.CreateOptions{},
	)
	return err
}

func CreateHeavyJob(command []string) error {
	job := newJob(command, HeavyJobCPURequest, HeavyJobMemRequest)
	_, err := K8sClient.BatchV1().Jobs("default").Create(
		context.Background(),
		job,
		metav1.CreateOptions{},
	)
	return err
}
