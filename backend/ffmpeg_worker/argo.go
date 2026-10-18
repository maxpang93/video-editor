package ffmpegworker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	wfv1 "github.com/argoproj/argo-workflows/v4/pkg/apis/workflow/v1alpha1"
	argoClient "github.com/argoproj/argo-workflows/v4/pkg/client/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type Segment struct {
	Start  int    `json:"start"`
	End    int    `json:"end"`
	Output string `json:"output"`
}

func SubmitTrimMergeWorkflow(
	ctx context.Context,
	config *rest.Config,
	videoPath string,
	segments []Segment,
) (string, error) {
	segmentsJSON, err := json.Marshal(segments)
	if err != nil {
		return "", fmt.Errorf("marshal workflow segments: %w", err)
	}

	parts := make([]string, len(segments))
	for i, segment := range segments {
		parts[i] = segment.Output
	}

	client, err := argoClient.NewForConfig(config)
	if err != nil {
		return "", fmt.Errorf("create Argo client: %w", err)
	}

	workflow := &wfv1.Workflow{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "trim-merge-"},
		Spec: wfv1.WorkflowSpec{
			WorkflowTemplateRef: &wfv1.WorkflowTemplateRef{Name: "trim-merge"},
			Arguments: wfv1.Arguments{
				Parameters: []wfv1.Parameter{
					{Name: "video", Value: wfv1.AnyStringPtr(videoPath)},
					{Name: "output", Value: wfv1.AnyStringPtr(GetOutputVideoPath(videoPath, 0))},
					{Name: "parts", Value: wfv1.AnyStringPtr(strings.Join(parts, " "))},
					{Name: "segments", Value: wfv1.AnyStringPtr(string(segmentsJSON))},
				},
			},
		},
	}

	created, err := client.ArgoprojV1alpha1().Workflows("argo").Create(ctx, workflow, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("submit trim-merge workflow: %w", err)
	}

	return created.Name, nil
}
