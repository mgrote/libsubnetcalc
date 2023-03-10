package subnets

import (
	"net"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCalculateSubnetsByHostCount(t *testing.T) {
	tests := []struct {
		description             string
		sourceNetCIDR           string
		requestedTotalHostCount int
		expectedSubnetCount     int
	}{
		{
			description:             "100.64.0.0/16 --> 1024 Hosts per subnet",
			sourceNetCIDR:           "100.64.0.0/16",
			requestedTotalHostCount: 1023,
			expectedSubnetCount:     64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			g := NewWithT(t)

			subnets, err := CalculateSubnetsByHostCount(tt.sourceNetCIDR, uint32(tt.requestedTotalHostCount))
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(len(subnets)).To(BeIdenticalTo(tt.expectedSubnetCount))
			for _, subnet := range subnets {
				g.Expect(tt.requestedTotalHostCount + 1).To(BeIdenticalTo(subnet.TotalHostsNum))
			}
		})
	}
}

func TestCalculateSubnetsByCIDR(t *testing.T) {
	tests := []struct {
		description         string
		sourceNetCIDR       string
		subnetCIDR          uint32
		expectedSubnetCount int
		expectedHostCount   int
	}{
		{
			description:         "100.64.0.0/16 --> /22 CIDR",
			sourceNetCIDR:       "100.64.0.0/16",
			subnetCIDR:          22,
			expectedSubnetCount: 64,
			expectedHostCount:   1024,
		},
		{
			description:         "100.64.0.0/16 --> /23 CIDR",
			sourceNetCIDR:       "100.64.0.0/16",
			subnetCIDR:          23,
			expectedSubnetCount: 128,
			expectedHostCount:   512,
		},
		{
			description:         "100.64.0.0/16 --> /24 CIDR",
			sourceNetCIDR:       "100.64.0.0/16",
			subnetCIDR:          24,
			expectedSubnetCount: 256,
			expectedHostCount:   256,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			g := NewWithT(t)

			subnets, err := CalculateSubnetsByCIDR(tt.sourceNetCIDR, tt.subnetCIDR)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(len(subnets)).To(BeIdenticalTo(tt.expectedSubnetCount))
			for _, subnet := range subnets {
				g.Expect(tt.expectedHostCount).To(BeIdenticalTo(subnet.TotalHostsNum))
			}
		})
	}
}

func TestCalculateSubnetsByCIDRWithRange(t *testing.T) {
	tests := []struct {
		description         string
		sourceNetCIDR       string
		subnetCIDR          uint32
		requestedSubnetNum  int
		expectedSubnetCount int
		expectedHostCount   int
	}{
		{
			description:         "100.64.0.0/16 --> /22 CIDR",
			sourceNetCIDR:       "100.64.0.0/16",
			subnetCIDR:          22,
			requestedSubnetNum:  50,
			expectedSubnetCount: 64,
			expectedHostCount:   1024,
		},
		{
			description:         "100.64.0.0/16 --> /23 CIDR",
			sourceNetCIDR:       "100.64.0.0/16",
			subnetCIDR:          23,
			requestedSubnetNum:  50,
			expectedSubnetCount: 128,
			expectedHostCount:   512,
		},
		{
			description:         "100.64.0.0/16 --> /24 CIDR",
			sourceNetCIDR:       "100.64.0.0/16",
			subnetCIDR:          24,
			requestedSubnetNum:  50,
			expectedSubnetCount: 256,
			expectedHostCount:   256,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			g := NewWithT(t)

			subnets, err := CalculateSubnetsByCIDR(tt.sourceNetCIDR, tt.subnetCIDR, tt.requestedSubnetNum)
			g.Expect(err).ShouldNot(HaveOccurred())
			g.Expect(len(subnets)).To(BeIdenticalTo(tt.requestedSubnetNum))
			for _, subnet := range subnets {
				g.Expect(tt.expectedHostCount).To(BeIdenticalTo(subnet.TotalHostsNum))
			}
		})
	}
}

func TestCalculateSubnetMaksFromAddressBits(t *testing.T) {
	tests := []struct {
		description             string
		potentialAddressPortion []uint32
		expectedNetMask         net.IPMask
		expectedTotalHosts      uint32
	}{
		{
			description:             "2 address bits set, smallest usable mask",
			potentialAddressPortion: []uint32{2, 3},
			expectedNetMask:         net.CIDRMask(30, 32),
			expectedTotalHosts:      4,
		},
		{
			description:             "3 address bits set",
			potentialAddressPortion: []uint32{4, 5, 6},
			expectedNetMask:         net.CIDRMask(29, 32),
			expectedTotalHosts:      8,
		},
		{
			description:             "4 address bits set",
			potentialAddressPortion: []uint32{8, 10, 15},
			expectedNetMask:         net.CIDRMask(28, 32),
			expectedTotalHosts:      16,
		},
		{
			description:             "8 address bits set",
			potentialAddressPortion: []uint32{128, 200, 255},
			expectedNetMask:         net.CIDRMask(24, 32),
			expectedTotalHosts:      256,
		},
		{
			description:             "10 address bits set",
			potentialAddressPortion: []uint32{512, 731, 1023},
			expectedNetMask:         net.CIDRMask(22, 32),
			expectedTotalHosts:      1024,
		},
	}
	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			g := NewWithT(t)

			for _, ap := range tt.potentialAddressPortion {
				netMask, totalHostCount := getSubnetMaskFromAddressBits(ap)
				expectedOnes, expectedBits := tt.expectedNetMask.Size()
				ones, bits := netMask.Size()
				g.Expect(expectedOnes).To(BeIdenticalTo(ones))
				g.Expect(expectedBits).To(BeIdenticalTo(bits))
				g.Expect(tt.expectedTotalHosts).To(BeIdenticalTo(totalHostCount))
			}
		})
	}
}
