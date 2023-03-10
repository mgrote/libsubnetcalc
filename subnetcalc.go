package subnets

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
)

type Subnet struct {
	NetworkCIDR   string
	Network       net.IPNet // TODO doubles IP and NetworkMask
	IP            net.IP
	NetworkMask   net.IPMask
	BroadcastIP   net.IP
	HostMinIP     net.IP
	HostMaxIP     net.IP
	HostsNum      int
	TotalHostsNum int
}

func CalculateSubnetsByCIDR(CIDRBlock string, cidr uint32, requestedSubnetCount ...int) ([]*Subnet, error) {
	// get subnet mask from cidr
	sourceNet, err := CalculateSubnet(CIDRBlock)
	if err != nil {
		return nil, err
	}
	subnetMask := net.CIDRMask(int(cidr), 32)
	totalHostCount := uint32(0xFFFFFFFF>>cidr + 1)

	return CalculateSubnets(sourceNet, subnetMask, totalHostCount, requestedSubnetCount...)
}

// CalculateSubnetsByHostCount
func CalculateSubnetsByHostCount(CIDRBlock string, hostNumber uint32, requestedSubnetCount ...int) ([]*Subnet, error) {
	sourceNet, err := CalculateSubnet(CIDRBlock)
	if err != nil {
		return nil, err
	}
	subnetMask, totalSubnetHosts := getSubnetMaskFromAddressBits(hostNumber)
	return CalculateSubnets(sourceNet, subnetMask, totalSubnetHosts, requestedSubnetCount...)
}

// CalculateSubnetsBySubnetCount divides a given CIDR block into a requested number of subnets.
func CalculateSubnetsBySubnetCount(CIDRBlock string, subnetNumber int, requestedSubnetCount ...int) ([]*Subnet, error) {
	sourceNet, err := CalculateSubnet(CIDRBlock)
	if err != nil {
		return nil, err
	}

	targetSubnetNum := uint32(float64(sourceNet.TotalHostsNum / subnetNumber))
	netMask, totalHosts := getSubnetMaskFromAddressBits(targetSubnetNum)
	return CalculateSubnets(sourceNet, netMask, totalHosts, requestedSubnetCount...)
}

// CalculateSubnets devides a given subnet in a range of subnets for the required count of contained hosts.
func CalculateSubnets(sourceNet *Subnet, subnetMask net.IPMask, totalSubnetHosts uint32, requestedSubnetCount ...int) ([]*Subnet, error) {
	expectedNetworkNum := int(float64(sourceNet.TotalHostsNum / int(totalSubnetHosts)))
	if len(requestedSubnetCount) > 0 {
		if expectedNetworkNum < requestedSubnetCount[0] {
			return nil, fmt.Errorf("requested subnet count %d exeeds maximal possible subnet count %d", requestedSubnetCount, expectedNetworkNum)
		}
		expectedNetworkNum = requestedSubnetCount[0]
	}

	maskOnes, subnetBits := subnetMask.Size()
	addressBits := subnetBits - maskOnes
	var subnets []*Subnet
	for i := 0; i < expectedNetworkNum; i++ {
		currentSubnetMask := i << addressBits
		currentSubnetIP := intToIP(ipToInt(sourceNet.IP) | uint32(currentSubnetMask))
		currentSubnet, err := CalculateSubnet(fmt.Sprintf("%s/%d", currentSubnetIP.String(), maskOnes))
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, currentSubnet)
	}
	return subnets, nil
}

// getSubnetMaskFromAddressBits delivers the IPMask and the minimal needed host count
// for any requested number of hosts contained by a requested subnet.
func getSubnetMaskFromAddressBits(addressBits uint32) (netMask net.IPMask, totalHostCount uint32) {
	maskBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(maskBytes, addressBits)

	networkMaskOnes := 0
	// search for first bit that is set
	for _, eightBits := range maskBytes {
		if eightBits == 0 {
			networkMaskOnes += 8
			continue
		}
		for eightBits&0x80 != 128 {
			networkMaskOnes++
			eightBits <<= 1
		}
		break
	}
	netMask = net.CIDRMask(networkMaskOnes, 32)
	totalHostCount = 0xFFFFFFFF>>networkMaskOnes + 1
	return netMask, totalHostCount
}

// GetHostIPsForSubnet calculates the IP addresses between the
// minimal and the maximal host address.
// The network address and the broadcast address are stripped.
func GetHostIPsForSubnet(CIDRBlock string) ([]net.IP, error) {
	ipnet, err := CalculateSubnet(CIDRBlock)
	if err != nil {
		return nil, err
	}
	host := ipToInt(ipnet.HostMinIP)
	lastHost := ipToInt(ipnet.HostMaxIP)

	IPs := make([]net.IP, ipnet.HostsNum)
	i := 0
	for host <= lastHost {
		currentIP := intToIP(host)
		IPs[i] = currentIP
		host++
		i++
	}
	return IPs, nil
}

func CalculateSubnet(CIDRBlock string) (*Subnet, error) {
	ipnet := Subnet{
		NetworkCIDR: CIDRBlock,
	}
	sourceNetStartIP, ipnetwork, err := net.ParseCIDR(CIDRBlock)
	if err != nil {
		return nil, err
	}

	ipnet.Network = *ipnetwork
	ipnet.IP = sourceNetStartIP
	ipnet.NetworkMask = ipnetwork.Mask

	// Convert IP bytes to int to allow bitwise operations.
	networkIPInt := ipToInt(sourceNetStartIP)

	// Mask with the host part bits for broadcast address.
	networkMaskOnes, _ := ipnetwork.Mask.Size()
	subnetIPOnes := 0xFFFFFFFF >> networkMaskOnes
	broadcastIPInt := networkIPInt | uint32(subnetIPOnes)

	hostMinIPInt := networkIPInt | 1
	hostMaxIPInt := broadcastIPInt &^ 1

	ipnet.TotalHostsNum = int(broadcastIPInt - networkIPInt + 1)
	ipnet.HostsNum = ipnet.TotalHostsNum - 2

	// Convert int back to bytes for regular net.IP.
	ipnet.BroadcastIP = intToIP(broadcastIPInt)
	ipnet.HostMinIP = intToIP(hostMinIPInt)
	ipnet.HostMaxIP = intToIP(hostMaxIPInt)
	return &ipnet, nil
}

func intToIP(intIP uint32) net.IP {
	IPBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(IPBytes, intIP)
	return IPBytes
}

func ipToInt(netIP net.IP) uint32 {
	return binary.BigEndian.Uint32(netIP.To4())
}

func (s *Subnet) String() string {
	return s.IP.String() + "/" + s.NetworkMask.String() + "\n" +
		"HostMin:     " + s.HostMinIP.String() + "\n" +
		"HostMax:     " + s.HostMaxIP.String() + "\n" +
		"Broadcast:   " + s.BroadcastIP.String() + "\n" +
		"Hosts:       " + strconv.Itoa(s.HostsNum) + "\n" +
		"Hosts total: " + strconv.Itoa(s.TotalHostsNum) + "\n"
}
