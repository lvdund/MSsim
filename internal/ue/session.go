package context

import (
	"fmt"
	"github.com/vishvananda/netlink"
	gnbContext "mssim/internal/gnb/context"
	"net"
)

type UEPDUSession struct {
	Id            uint8
	GnbPduSession *gnbContext.GnbPDUSession
	ueIP          string
	ueGnbIP       net.IP
	tun           netlink.Link
	rule          *netlink.Rule
	routeTun      *netlink.Route
	vrf           *netlink.Vrf
	stopSignal    chan bool
	Wait          chan bool
	T3580Retries  int

	Exp Experiment

	// TS 24.501 - 6.1.3.2.1.1 State Machine for Session Management
	StateSM int
}

func (pduSession *UEPDUSession) SetIp(ip []uint8) {
	pduSession.ueIP = fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
}

func (pduSession *UEPDUSession) GetIp() string {
	return pduSession.ueIP
}

func (pduSession *UEPDUSession) SetGnbIp(ip net.IP) {
	pduSession.ueGnbIP = ip
}

func (pduSession *UEPDUSession) GetGnbIp() net.IP {
	return pduSession.ueGnbIP
}

func (pduSession *UEPDUSession) SetStopSignal(stopSignal chan bool) {
	pduSession.stopSignal = stopSignal
}

func (pduSession *UEPDUSession) GetStopSignal() chan bool {
	return pduSession.stopSignal
}

func (pduSession *UEPDUSession) GetPduSesssionId() uint8 {
	return pduSession.Id
}

func (pduSession *UEPDUSession) SetTunInterface(tun netlink.Link) {
	pduSession.tun = tun
}

func (pduSession *UEPDUSession) GetTunInterface() netlink.Link {
	return pduSession.tun
}

func (pduSession *UEPDUSession) SetTunRule(rule *netlink.Rule) {
	pduSession.rule = rule
}

func (pduSession *UEPDUSession) GetTunRule() *netlink.Rule {
	return pduSession.rule
}

func (pduSession *UEPDUSession) SetTunRoute(route *netlink.Route) {
	pduSession.routeTun = route
}

func (pduSession *UEPDUSession) GetTunRoute() *netlink.Route {
	return pduSession.routeTun
}

func (pduSession *UEPDUSession) SetVrfDevice(vrf *netlink.Vrf) {
	pduSession.vrf = vrf
}

func (pduSession *UEPDUSession) GetVrfDevice() *netlink.Vrf {
	return pduSession.vrf
}

func (pdu *UEPDUSession) SetStateSM_PDU_SESSION_INACTIVE() {
	pdu.StateSM = SM5G_PDU_SESSION_INACTIVE
}

func (pdu *UEPDUSession) SetStateSM_PDU_SESSION_ACTIVE() {
	pdu.StateSM = SM5G_PDU_SESSION_ACTIVE
}

func (pdu *UEPDUSession) SetStateSM_PDU_SESSION_PENDING() {
	pdu.StateSM = SM5G_PDU_SESSION_ACTIVE_PENDING
}

func (pduSession *UEPDUSession) GetStateSM() int {
	return pduSession.StateSM
}
