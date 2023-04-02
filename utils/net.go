package utils

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func CIDRRangeToIPv4Range(cidrs []string) (ipStart string, ipEnd string, err error) {

	var ip uint32 // ip address

	var ipS uint32 // Start IP address range
	var ipE uint32 // End IP address range

	for _, CIDR := range cidrs {

		cidrParts := strings.Split(CIDR, "/")

		ip = iPv4ToUint32(cidrParts[0])
		bits, _ := strconv.ParseUint(cidrParts[1], 10, 32)

		if ipS == 0 || ipS > ip {
			ipS = ip
		}

		ip = ip | (0xFFFFFFFF >> bits)

		if ipE < ip {
			ipE = ip
		}
	}

	ipStart = uInt32ToIPv4(ipS)
	ipEnd = uInt32ToIPv4(ipE)

	return ipStart, ipEnd, err
}

// Convert IPv4 range into CIDR
func IPv4RangeToCIDRRange(ipStart string, ipEnd string) (cidrs []string, err error) {

	cidr2mask := []uint32{
		0x00000000, 0x80000000, 0xC0000000,
		0xE0000000, 0xF0000000, 0xF8000000,
		0xFC000000, 0xFE000000, 0xFF000000,
		0xFF800000, 0xFFC00000, 0xFFE00000,
		0xFFF00000, 0xFFF80000, 0xFFFC0000,
		0xFFFE0000, 0xFFFF0000, 0xFFFF8000,
		0xFFFFC000, 0xFFFFE000, 0xFFFFF000,
		0xFFFFF800, 0xFFFFFC00, 0xFFFFFE00,
		0xFFFFFF00, 0xFFFFFF80, 0xFFFFFFC0,
		0xFFFFFFE0, 0xFFFFFFF0, 0xFFFFFFF8,
		0xFFFFFFFC, 0xFFFFFFFE, 0xFFFFFFFF,
	}

	ipStartUint32 := iPv4ToUint32(ipStart)
	ipEndUint32 := iPv4ToUint32(ipEnd)

	if ipStartUint32 > ipEndUint32 {
		log.Fatalf("start IP:%s must be less than end IP:%s", ipStart, ipEnd)
	}

	for ipEndUint32 >= ipStartUint32 {
		maxSize := 32
		for maxSize > 0 {

			maskedBase := ipStartUint32 & cidr2mask[maxSize-1]

			if maskedBase != ipStartUint32 {
				break
			}
			maxSize--

		}

		x := math.Log(float64(ipEndUint32-ipStartUint32+1)) / math.Log(2)
		maxDiff := 32 - int(math.Floor(x))
		if maxSize < maxDiff {
			maxSize = maxDiff
		}

		cidrs = append(cidrs, uInt32ToIPv4(ipStartUint32)+"/"+strconv.Itoa(maxSize))

		ipStartUint32 += uint32(math.Exp2(float64(32 - maxSize)))
	}

	return cidrs, err
}

// Convert IPv4 to uint32
func iPv4ToUint32(iPv4 string) uint32 {

	ipOctets := [4]uint64{}

	for i, v := range strings.SplitN(iPv4, ".", 4) {
		ipOctets[i], _ = strconv.ParseUint(v, 10, 32)
	}

	result := (ipOctets[0] << 24) | (ipOctets[1] << 16) | (ipOctets[2] << 8) | ipOctets[3]

	return uint32(result)
}

// Convert uint32 to IP
func uInt32ToIPv4(iPuInt32 uint32) (iP string) {
	iP = fmt.Sprintf("%d.%d.%d.%d",
		iPuInt32>>24,
		(iPuInt32&0x00FFFFFF)>>16,
		(iPuInt32&0x0000FFFF)>>8,
		iPuInt32&0x000000FF)
	return iP
}

// Parse 'ips' parameter into the array of CDIR (https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing)
func GetCIDRs(ipsParameter string) (cidrs []string, err error) {

	paramParts := strings.Split(ipsParameter, ",")

	var cidrRegEx = regexp.MustCompile(`^([0-9]{1,3}\.){3}[0-9]{1,3}(\/([0-9]|[1-2][0-9]|3[0-2]))?$`)

	for _, v := range paramParts {
		paramParts := strings.TrimSpace(v)
		log.Println(cidrRegEx.MatchString(paramParts))
		if cidrRegEx.MatchString(paramParts) {
			cidrs = append(cidrs, paramParts)
			continue
		}

		ipParamParts := strings.Split(paramParts, "-")

		ipStart := strings.TrimSpace(ipParamParts[0])
		ipEnd := ipStart
		if len(ipParamParts) > 1 {
			ipEnd = strings.TrimSpace(ipParamParts[1])
		}

		paramPartsCidrs, err := IPv4RangeToCIDRRange(ipStart, ipEnd)

		if err != nil {
			fmt.Errorf("enable to parse IP range: %s - %s", ipStart, ipEnd)
		}

		cidrs = append(cidrs, paramPartsCidrs...)
	}

	return cidrs, err
}
