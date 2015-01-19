package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	EtcdConn            *etcd.Client
	DockerConn          *docker.Client
	cli_docker_endpoint string
	etcd_nodes          []string
	cli_containers      []string
	cli_etcd_nodes      string
	basekey             string
	wg                  sync.WaitGroup
	cli_all             bool
)

func init() {
	flag.StringVar(&cli_docker_endpoint, "d", "unix:///var/run/docker.sock", "[d]ocker endpoint")
	flag.StringVar(&cli_etcd_nodes, "e", "http://127.0.0.1:4001", "Comma separated list of [e]tcd endpoints")
	flag.StringVar(&basekey, "k", "/test", "etcd base [k]ey")
	flag.BoolVar(&cli_all, "a", false, "Update [a]ll containers")

	flag.Usage = func() {
		fmt.Printf("\n---------------------\nUsage: ip2etcd [options] containers\nVersion: 0.2\n---------------------\n")
		flag.PrintDefaults()
		fmt.Printf("---------------------\n")
	}

	flag.Parse()

	cli_containers = flag.Args()

	for _, node := range strings.Split(cli_etcd_nodes, ",") {
		etcd_nodes = append(etcd_nodes, node)
	}
}

func main() {
	etcdClient := etcd.NewClient(etcd_nodes)
	dockerClient, _ := docker.NewClient(cli_docker_endpoint)

	if cli_all {
		//updateEtcd(dockerClient, etcdClient)
		containerIds := getContainerIds(dockerClient)
		containerMap := getContainerMap(dockerClient)
		for _, container := range containerIds {
			shortId := trimContainerId(container)
			name := containerMap[shortId]
			var key string
			if len(name) > 0 {
				key = basekey + name + "/ip"
			} else {
				key = basekey + shortId + "/ip"
			}
			updateContainer(container, dockerClient, etcdClient, key)
		}
		os.Exit(0)
	}

	for _, cli_arg := range cli_containers {
		key := basekey + "/" + strings.Trim(cli_arg, " ") + "/ip"
		updateContainer(cli_arg, dockerClient, etcdClient, key)
	}
}

func idOrName(client *docker.Client, idorname string) string {
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.Println("Error getting container info! ")
		return ""
	}
	for _, container := range containers {
		for _, name := range container.Names {
			if name == "/"+idorname {
				return container.ID
			}
		}
		if strings.Contains(container.ID, idorname) {
			return container.ID
		}
	}
	return ""
}

func updateContainer(nameOrId string, dclient *docker.Client, eclient *etcd.Client, etcdKey string) {
	containerId := idOrName(dclient, nameOrId)
	shortname := trimContainerId(nameOrId)
	if len(containerId) <= 0 {
		log.Printf("Container %s does not exist!\n", shortname)
		return
	}
	container, err := dclient.InspectContainer(containerId)
	var ip string
	if err != nil {
		ip = ""
	} else {
		ip = container.NetworkSettings.IPAddress
	}
	if len(ip) <= 0 {
		log.Printf("No IP address found for container %s, skipping...\n", shortname)
		return
	}
	log.Printf("Found IP %s for container %s\n", ip, shortname)

	currentValue := getKey(eclient, etcdKey)
	if currentValue == ip {
		log.Printf("Current value matches proposed update, no change needed!\n")
		return
	}
	if currentValue != "" {
		log.Printf("Found current value for %s: %s.  Updating...\n", etcdKey, currentValue)
	}
	err = setKey(eclient, etcdKey, ip)
	if err != nil {
		log.Println("Error setting new value! Error text: ", err.Error())
		return
	}
	log.Printf("Successfully set %s as new value for %s.", ip, etcdKey)
}

func getContainerMap(client *docker.Client) map[string]string {
	var m = make(map[string]string)
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.Printf("Encountered error getting container info!\n%s", err.Error())
	}
	for _, container := range containers {
		shortId := trimContainerId(container.ID)
		concatNames := fmt.Sprintf("%s", strings.Join(container.Names, ":"))
		m[shortId] = concatNames
	}
	return m
}

func trimContainerId(id string) string {
	re := regexp.MustCompile("^[A-Za-z0-9\\/]{1,12}")
	return fmt.Sprintf("%s", re.FindString(id))
}

func getContainerIds(client *docker.Client) []string {
	var results []string
	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.Printf("Encountered error getting container IDs:\n%s", err.Error())
	}
	for _, container := range containers {
		results = append(results, container.ID)
	}
	return results
}

func getKey(client *etcd.Client, key string) string {
	resp, err := client.Get(key, false, false)
	if err != nil {
		return ""
	}
	return resp.Node.Value
}

func setKey(client *etcd.Client, key string, value string) error {
	_, err := client.Set(key, value, 0)
	return err
}
