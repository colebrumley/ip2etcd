package main

import (
	"flag"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"strings"
)

var (
	EtcdConn            *etcd.Client
	DockerConn          *docker.Client
	cli_docker_endpoint string
	etcd_nodes          []string
	cli_container       string
	cli_etcd_nodes      string
	key                 string
	quiet               bool
)

func init() {
	const (
		default_etcd_nodes      = "http://127.0.0.1:4001"
		node_usage              = "Comma separated list of etcd nodes"
		default_etcd_key        = "/net"
		key_usage               = "Etcd base key"
		default_docker_endpoint = "unix:///var/run/docker.sock"
		endpoint_usage          = "Docker socket or URL"
		default_container       = ""
		container_usage         = "Target container"
	)
	flag.StringVar(&cli_container, "container", default_container, container_usage)
	flag.StringVar(&cli_container, "c", default_container, container_usage+" (Shorthand)")
	flag.StringVar(&cli_docker_endpoint, "docker-endpoint", default_docker_endpoint, endpoint_usage)
	flag.StringVar(&cli_docker_endpoint, "d", default_docker_endpoint, endpoint_usage+" (Shorthand)")
	flag.StringVar(&cli_etcd_nodes, "etcd-nodes", default_etcd_nodes, node_usage)
	flag.StringVar(&cli_etcd_nodes, "e", default_etcd_nodes, node_usage+" (Shorthand)")
	flag.StringVar(&key, "key", default_etcd_key, key_usage)
	flag.StringVar(&key, "k", default_etcd_key, key_usage+" (Shorthand)")
	flag.BoolVar(&quiet, "q", false, "Don't error when container IP is null")

	flag.Usage = func() {
		fmt.Printf("\n---------------------\nUsage: ip2etcd [options]\nVersion: 0.1\n---------------------\n")
		flag.PrintDefaults()
		fmt.Printf("---------------------\n")
	}

	flag.Parse()

	for _, node := range strings.Split(cli_etcd_nodes, ",") {
		etcd_nodes = append(etcd_nodes, node)
	}

	EtcdConn = etcd.NewClient(etcd_nodes)
	DockerConn, _ = docker.NewClient(cli_docker_endpoint)
	key = key + "/" + cli_container + "/ip"
}

func main() {
	if len(cli_container) <= 0 {
		log.Println("Must supply the container!")
		os.Exit(1)
	}
	container_id := id_or_name(cli_container)
	if len(container_id) <= 0 {
		log.Println("No container with name/id ", cli_container)
		os.Exit(1)
	}
	//log.Printf("Found ID %s for container %s\n", container_id, cli_container)
	ip := get_container_ip(cli_container)
	if len(ip) <= 0 && !quiet {
		log.Println("No IP address found!  Nothing to set, erroring out...")
		os.Exit(1)
	} else if len(ip) <= 0 && quiet {
		log.Printf("No IP address found!  I have nothing to do, but you told me to shut up so we're all good here.\n")
		os.Exit(0)
	}
	log.Printf("Found IP %s for container %s\n", ip, cli_container)
	current_value := get_key(key)
	if current_value == ip {
		log.Printf("Current value matches proposed update, no change needed!\n")
		os.Exit(0)
	}
	if current_value != "" {
		log.Printf("Found current value for %s: %s.  Updating...\n", key, current_value)
	}
	err := set_key(key, ip)
	if err != nil {
		log.Println("Error setting new value! Error text: ", err.Error())
		os.Exit(1)
	}
	log.Printf("Successfully set %s as new value for %s, exiting...", ip, key)
}

func get_key(key string) string {
	resp, err := EtcdConn.Get(key, false, false)
	if err != nil {
		//log.Printf("Etcd client error getting %s!  Error text:%s\n", key, err.Error())
		return ""
	}
	return resp.Node.Value
}

func set_key(key string, value string) error {
	_, err := EtcdConn.Set(key, value, 0)
	return err
}

func id_or_name(idorname string) string {
	containers, err := DockerConn.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil {
		log.Fatal("Error getting container info!: ", err.Error())
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

func get_container_ip(id string) string {
	container, err := DockerConn.InspectContainer(id)
	if err != nil {
		log.Printf("Encountered error getting info on container %s!\nError text:\n%s", id, err.Error())
	}
	return container.NetworkSettings.IPAddress
}
