package main

import (
	"fmt"
	"log"
	"os"
)

// Example usage: go run main.go kubeclient.go "$(oc whoami -t)" myuser myspace
func main() {
	url := "https://openshift.io/_p/oso"

	if len(os.Args) < 4 {
		log.Fatalln("usage: go run main.go kubeclient.go kube_token user_namespace space_name")
	}

	token := os.Args[1]
	userNamespace := os.Args[2]
	spaceName := os.Args[3]

	kc, err := NewKubeClient(url, token, userNamespace)
	if err != nil {
		log.Fatalln(err)
	}

	space, err := kc.GetSpace(spaceName)
	if err != nil {
		log.Fatalln(err)
	}
	for _, appn := range space.Applications {
		fmt.Println("Application:", *appn.Name)
		for _, env := range appn.Pipeline {
			fmt.Println("\tEnvironment:", *env.Name)
			fmt.Println("\t\tPodsStarting:", *env.Stats.Pods.Starting)
			fmt.Println("\t\tPodsRunning:", *env.Stats.Pods.Running)
			fmt.Println("\t\tPodsStopping:", *env.Stats.Pods.Stopping)
		}
	}

	envs, err := kc.GetEnvironments()
	if err != nil {
		log.Fatalln(err)
	}
	for _, env := range envs {
		fmt.Println("Environment:", *env.Name)
		fmt.Println("\tCPU Used:", *env.Quota.Cpucores.Used)
		fmt.Println("\tCPU Limit:", *env.Quota.Cpucores.Quota)
		fmt.Println("\tMemory Used:", *env.Quota.Memory.Used)
		fmt.Println("\tMemory Limit:", *env.Quota.Memory.Quota)
	}
}
