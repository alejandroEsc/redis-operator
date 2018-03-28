package redis

import (
	api "gitlab.com/mvenezia/redis-operator/pkg/apis/redis/v1alpha1"
	"k8s.io/client-go/kubernetes"
	"gitlab.com/mvenezia/redis-operator/pkg/client/clientset/versioned"

	"time"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log"
	yaml "gopkg.in/yaml.v2"
	"os/exec"
)

var (
	reconcileInterval         = 8 * time.Second
	podTerminationGracePeriod = int64(5)
)

type redisEventType string

const (
	eventModifyRedis redisEventType = "Modify"
)

type redisEvent struct {
	typ     redisEventType
	cluster *api.Redis
}

type Config struct {
	ServiceAccount string

	KubeCli   kubernetes.Interface
	RedisCRCli versioned.Interface
}

type Redis struct {
	logger *logrus.Entry

	config Config

	redis *api.Redis

	status api.RedisStatus

	eventCh chan *redisEvent
	stopCh chan struct{}

	eventsCli corev1.EventInterface
}

func New(config Config, cl *api.Redis) *Redis {
	lg := logrus.WithField("pkg", "redis").WithField("redis-name", cl.Name)

	c := &Redis{
		logger:    lg,
		config:    config,
		redis:   cl,
		eventCh:   make(chan *redisEvent, 100),
		stopCh:    make(chan struct{}),
		status:    *(cl.Status.DeepCopy()),
		eventsCli: config.KubeCli.Core().Events(cl.Namespace),
	}

	log.Printf("Adding Redis Instance %s\n", cl.ObjectMeta.Name)
	command := "helm"
	arguments := []string{"install", "/samsung/go/src/gitlab.com/mvenezia/redis-operator/assets/redis-ha", "--name", cl.ObjectMeta.Name + "-redis", "--namespace", cl.ObjectMeta.Namespace }
	cmdOut, err := exec.Command(command, arguments...).Output()
	if err != nil {
		log.Printf("Error executing command: %s\n", err)
	}
	log.Printf("Helm response: %s\n", cmdOut)

	return c
}

func (c *Redis) Update(cl *api.Redis) {

	log.Printf("Modifying Redis Instance %s\n", cl.ObjectMeta.Name)

}

func (c *Redis) Delete(cl *api.Redis) {

	log.Printf("Deleting Redis Instance %s\n", cl.ObjectMeta.Name)
	command := "helm"
	arguments := []string{"delete", "--purge", cl.ObjectMeta.Name + "-redis" }
	cmdOut, err := exec.Command(command, arguments...).Output()
	if err != nil {
		log.Printf("Error executing command: %s\n", err)
	}
	log.Printf("Helm response: %s\n", cmdOut)

}

func generateYaml(cl *api.Redis) ([]byte, error) {

	vals, err := yaml.Marshal(map[string]interface{}{
		"replicas.servers": 	cl.Spec.Redis.Replicas,
		"replicas.sentinels":   cl.Spec.Sentinel.Replicas,
	})

	return vals, err
}