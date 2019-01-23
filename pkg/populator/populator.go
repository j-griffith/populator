package populator

import (
	"fmt"
	"log"

	"github.com/j-griffith/populator/pkg/api/types/v1alpha1"
	batch "k8s.io/api/batch/v1"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ttl = int32(30)

// JobRequest encapsulates all the details we need to run a populator job
type JobRequest struct {
	Name       string
	Image      string
	Args       []string
	MountPoint string
	PVCName    string
}

// CreateJobRequest is a helper function to take a pvc and a populator object and set up a JobRequest that caller can then use to launch the populator job.
// For users that want to roll their own JobRequst and call RunPopulatorJob directly they can ignore this function
func CreateJobRequest(pvc *core_v1.PersistentVolumeClaim, p *v1alpha1.Populator) error {
	switch p.Spec.Type {
	case "git":
		log.Printf("handle git type for populator: %v", p.Spec)
	case "s3":
		log.Printf("Sorry, not implemented yet")
	default:
		log.Printf("sorry, I don't know what to do with the type: %s", p.Spec.Type)
		return fmt.Errorf("unknown Populator Type (%s)", p.Spec.Type)

	}
	return nil
}

// RunPopulatorJob takes a K8s Client and a JobRequest and uses it to build a jobSpec, and launch the job.  We return the name of the Job to the caller
// The aim here is to have a pretty generic template for the various types of populators, and we can just differentiate by the image
// specified and the args supplied
func RunPopulatorJob(r *JobRequest) {
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-job",
		},
		Spec: batch.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "git-populator",
					},
				},
				Spec: core_v1.PodSpec{
					Containers: []core_v1.Container{
						{
							Name:  "git-pop",
							Image: "jgriffith/my-gitpop",
							Args:  []string{"a", "b", "c"},
							VolumeMounts: []core_v1.VolumeMount{
								{
									Name:      r.PVCName,
									MountPath: "/" + r.MountPoint,
								},
							},
						},
					},
					RestartPolicy: core_v1.RestartPolicyOnFailure,
					Volumes: []core_v1.Volume{
						{
							Name: "git-pop",
							VolumeSource: core_v1.VolumeSource{
								PersistentVolumeClaim: &core_v1.PersistentVolumeClaimVolumeSource{
									ClaimName: r.PVCName,
									ReadOnly:  false,
								},
							},
						},
					},
				},
			},
		},
	}
	log.Printf("Our job is: %v", job)

}
