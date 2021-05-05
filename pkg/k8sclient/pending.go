package k8sclient

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *K8sClient) WaitUntilNoPodsPending() error {
	// TODO: consider WaitGroup and go-routine, instead of sleeping on
	// main thread
	for {
		pods, err := c.Client.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return err
		}
		pending := 0
		for _, pod := range pods.Items {
			podCpy := pod
			phase := getPodPhase(&podCpy)
			if phase == "Pending" {
				fmt.Println(pod.ObjectMeta.Name)
				pending++
			}
		}
		if pending == 0 {
			return nil
		}
		fmt.Printf("%v pods pending, sleeping\n", pending)
		time.Sleep(15 * time.Second)
	}
	return nil
}
