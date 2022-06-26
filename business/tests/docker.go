package tests

import (
	"bytes"
	"encoding/json"
	"net"
	"os/exec"
	"testing"
)

type Container struct {
	ID   string
	Host string // IP:Port
}

func startContainer(t *testing.T, image string, port string, args ...string) *Container {
	arg := []string{"run", "-P", "-d"}
	arg = append(arg, args...)
	arg = append(arg, image)

	cmd := exec.Command("docker", arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("couldn't start container %s: %v", image, err)
	}

	id := out.String()[:12]

	cmd = exec.Command("docker", "inspect", id)
	out.Reset()
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("couldn't inspect container %s: %v", id, err)
	}

	var doc []map[string]any
	if err := json.Unmarshal(out.Bytes(), &doc); err != nil {
		t.Fatalf("couldn't decode json: %v", err)
	}

	ip, randPort := extractIPPort(t, doc, port)

	c := Container{
		ID:   id,
		Host: net.JoinHostPort(ip, randPort),
	}

	t.Logf("Image:       %s", image)
	t.Logf("ContainerID: %s", c.ID)
	t.Logf("Host:        %s", c.Host)

	return &c
}

func stopContainer(t *testing.T, id string) {
	if err := exec.Command("docker", "stop", id).Run(); err != nil {
		t.Fatalf("couldn't stop container: %v", err)
	}
	t.Logf("Stopped: %s", id)

	if err := exec.Command("docker", "rm", id, "-v").Run(); err != nil {
		t.Fatalf("couldn't remove container: %v", err)
	}
	t.Logf("Removed: %s", id)
}

func dumpContainerLogs(t *testing.T, id string) {
	out, err := exec.Command("docker", "logs", id).CombinedOutput()
	if err != nil {
		t.Fatalf("couldn't log container: %v", err)
	}
	t.Logf("Logs for %s:\n%s", id, out)
}

func extractIPPort(t *testing.T, doc []map[string]any, port string) (string, string) {
	nw, exists := doc[0]["NetworkSettings"]
	if !exists {
		t.Fatalf("couldn't get network settings")
	}

	ports, exists := nw.(map[string]any)["Ports"]
	if !exists {
		t.Fatalf("couldn't get network ports settings")
	}

	tcp, exists := ports.(map[string]any)[port+"/tcp"]
	if !exists {
		t.Fatalf("couldn't get network ports/tcp settings")
	}

	list, exists := tcp.([]any)
	if !exists {
		t.Fatalf("couldn't get network ports/tcp list settings")
	}
	if len(list) != 1 {
		t.Fatalf("couldn't get network ports/tcp list settings")
	}

	data, exists := list[0].(map[string]any)
	if !exists {
		t.Fatalf("couldn't get network ports/tcp list data")
	}

	return data["HostIp"].(string), data["HostPort"].(string)
}
