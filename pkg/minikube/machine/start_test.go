package machine

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
)

const initialEtcHostsContent string = `127.0.0.1	localhost
127.0.1.1	mymachinename

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
127.0.0.1	host.minikube.internal
192.168.42.139	control-plane.minikube.internal
10.8.0.22	control-plane.minikube.internal
`

const expectedEtcHostsContent string = `127.0.0.1	localhost
127.0.1.1	mymachinename

# The following lines are desirable for IPv6 capable hosts
::1     ip6-localhost ip6-loopback
fe00::0 ip6-localnet
127.0.0.1	host.minikube.internal
10.1.2.3	control-plane.minikube.internal
`

func TestAddHostAliasInner(t *testing.T) {
	// Arrange
	tempFilePath, err := writeContentToTempFile(initialEtcHostsContent)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tempFilePath)
	hostname := "control-plane.minikube.internal"
	ip := net.ParseIP("10.1.2.3")
	hostsFileLine := fmt.Sprintf("%s\t%s", ip, hostname)

	// Act
	cmd := addHostAliasCommand(hostname, hostsFileLine, false, tempFilePath)
	if err = cmd.Run(); err != nil {
		t.Error(err)
	}

	// Assert
	if err != nil {
		t.Error(err)
	}

	buff, err := ioutil.ReadFile(tempFilePath)
	if err != nil {
		t.Error(err)
	}

	actualEtcHostsContent := string(buff)
	if expectedEtcHostsContent != actualEtcHostsContent {
		t.Errorf("\n=== EXPECTED ===\n%s\n=== ACTUAL ===\n%s\n",
			expectedEtcHostsContent,
			actualEtcHostsContent)
	}
}

func writeContentToTempFile(content string) (string, error) {
	etcHosts, err := ioutil.TempFile("", "hosts")
	if err != nil {
		return "", err
	}

	path := etcHosts.Name()
	if _, err = etcHosts.WriteString(content); err != nil {
		return "", err
	}

	if err = etcHosts.Close(); err != nil {
		return "", err
	}

	return path, nil
}
