package populator

import (
	"fmt"
	"log"

	"github.com/j-griffith/populator/pkg/api/types/v1alpha1"
	batch "k8s.io/api/batch/v1"
	core_v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GitPopulatorImage is the provided container image to handle git population
// the default git populator is pretty simple, it's entrypoint is a simple script
// to clone <branch> <repo> <destination-folder>
const GitPopulatorImage = "jgriffith/git-populator"

var ttl = int32(30) // our default ttl for completed containers is 30 seconds

// JobRequest encapsulates all the details we need to run a populator job
type JobRequest struct {
	Name       string
	Image      string
	Args       []string
	MountPoint string
	PVCName    string
}

// CreateJobFromObjects is a helper function to take a pvc and a populator object and set up a JobRequest that caller can then use to launch the populator job.
// For users that want to roll their own JobRequst and call RunPopulatorJob directly they can ignore this function
func CreateJobFromObjects(c kubernetes.Interface, pvc *core_v1.PersistentVolumeClaim, p *v1alpha1.Populator) (*batch.Job, error) {
	var job *batch.Job
	req := &JobRequest{}

	switch p.Spec.Type {
	case "git":
		log.Printf("creating job for git-populator: %v", p.Spec)
		req = &JobRequest{
			Name:       p.GetObjectMeta().GetName() + "-pvc-" + pvc.Name,
			Image:      GitPopulatorImage,
			MountPoint: p.Spec.Mountpoint,
			PVCName:    pvc.Name,
			Args:       []string{p.Spec.Git.Repo, p.Spec.Git.Branch, p.Spec.Mountpoint},
		}
		job = BuildJobSpec(req)
	case "s3":
		log.Printf("Sorry, not implemented yet")
	default:
		log.Printf("sorry, I don't know what to do with the type: %s", p.Spec.Type)
		return nil, fmt.Errorf("unknown Populator Type (%s)", p.Spec.Type)
	}

	job = BuildJobSpec(req)
	return RunPopulatorJob(c, job, pvc.Namespace)
}

// BuildJobSpec takes a JobRequest and uses it to build a jobSpec, and launch the job.  We return the name of the Job to the caller
// The aim here is to have a pretty generic template for the various types of populators, and we can just differentiate by the image
// specified and the args supplied, we also make this public so users can choose to call it without using a formal populator object
func BuildJobSpec(r *JobRequest) *batch.Job {
	job := &batch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: batch.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "populator",
					},
				},
				Spec: core_v1.PodSpec{
					Containers: []core_v1.Container{
						{
							Name:  r.Name,
							Image: r.Image,
							Args:  r.Args,
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
							Name: r.PVCName,
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
	return job
}

// RunPopulatorJob kicks off a Kubernetes Job using the supplied k8s client, and Job Spec
func RunPopulatorJob(c kubernetes.Interface, j *batch.Job, namespace string) (*batch.Job, error) {
	jobClient := c.Batch().Jobs(namespace)
	result, err := jobClient.Create(j)
	if err != nil {
		log.Printf("error encountered launching job %s, %v", j.Name, err)
	}
	return result, err
}
