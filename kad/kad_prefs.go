package kad

import (
	"hahajing/com"
	"net"
)

const localUDPPort = 1979
const localTCPPort = 1988

// Prefs is my preferences.
type Prefs struct {
	kadID ID

	tcpPort uint16 // useless only for sending packet to client

	udpKey uint32 // used to generate my verify key with destination IP for sending packet. Please note difference with struct UDPKey.

	externIP uint32

	localIP      uint32
	localUDPPort uint16 // used to UDP connection listen, if not firewalled, it's same as @externUDPPort
}

func (p *Prefs) start() {
	p.kadID.generate()
	p.udpKey = random32()
	p.tcpPort = localTCPPort

	p.localUDPPort = localUDPPort

	p.initLocalIP()
}

func (p *Prefs) getUDPVerifyKey(targetIP uint32) uint32 {
	ui64Buffer := uint64(p.udpKey)
	ui64Buffer <<= 32
	ui64Buffer |= uint64(targetIP)

	md5 := Md5Sum{}
	md5.calculate(uint64ToByte(ui64Buffer))

	rawHash := md5.getRawHash()
	ui32Hash := byteToUint32Slice(rawHash)

	key := ui32Hash[0]
	for _, hash := range ui32Hash[1:] {
		key ^= hash
	}
	return key%0xFFFFFFFE + 1
}

func (p *Prefs) getKadID() *ID {
	return &p.kadID
}

func (p *Prefs) getPublicIP() uint32 {
	return p.externIP
}

func (p *Prefs) getTCPPort() uint16 {
	return p.tcpPort
}

// Get preferred outbound ip of this machine
func (p *Prefs) initLocalIP() {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		com.HhjLog.Critical(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	com.HhjLog.Infof("Local outbound IP: %s\n", localAddr.IP.String())

	p.localIP = ip2I(localAddr.IP)
}
