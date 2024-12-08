package context

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/reogac/nas"
	"mssim/config"
	gnbContext "mssim/internal/gnb/context"
	"mssim/internal/sec"
	"sync"
	"time"

	"github.com/reogac/sbi/models"
	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
)

// 5GMM main states in the UE.
const MM5G_NULL = 0x00
const MM5G_DEREGISTERED = 0x01
const MM5G_REGISTERED_INITIATED = 0x02
const MM5G_REGISTERED = 0x03
const MM5G_SERVICE_REQ_INIT = 0x04
const MM5G_DEREGISTERED_INIT = 0x05
const MM5G_IDLE = 0x06

// 5GSM main states in the UE.
const SM5G_PDU_SESSION_INACTIVE = 0x00
const SM5G_PDU_SESSION_ACTIVE_PENDING = 0x01
const SM5G_PDU_SESSION_ACTIVE = 0x02

type ScenarioMessage struct {
	StateChange int
}

type AuthContext struct {
	kamf     []byte
	rand     []byte
	ngKsi    nas.KeySetIdentifier
	sqn      string
	amf      string
	milenage *sec.Milenage
}

type UeContext struct {
	id uint8

	StateMM           int
	gnbInboundChannel chan gnbContext.UEMessage
	gnbRx             chan gnbContext.UEMessage
	gnbTx             chan gnbContext.UEMessage
	drx               *time.Ticker
	PduSession        [16]*UEPDUSession

	secCap *nas.UeSecurityCapability
	supi   string
	msin   string
	snn    string
	suci   nas.MobileIdentity
	guti   *nas.Guti
	nasPdu []byte //registration request

	auth   AuthContext          //on-going authentication context
	secCtx *nas.SecurityContext //current security context

	// TODO: Modify config so you can configure these parameters per PDUSession
	Dnn        string
	Snssai     models.Snssai
	TunnelMode config.TunnelMode

	// Sync primitive
	scenarioChan chan ScenarioMessage

	ExpFile string
	lock    sync.Mutex
}

func CreateUe(conf config.Config, id int, ueMgrChannel chan UeTesterMessage, gnbInboundChannel chan gnbContext.UEMessage, wg *sync.WaitGroup, logFile string) chan ScenarioMessage {
	scenarioChan := make(chan ScenarioMessage)
	ue := &UeContext{
		id:     uint8(id),
		msin:   conf.Ue.Msin,
		secCap: conf.Ue.GetUESecurityCapability(),
	}
	//integAlg, cipherAlg := auth.SelectAlgorithms(ue.UeSecurity.UeSecurityCapability)
	op, _ := hex.DecodeString("0xc9e8763286b5b9ffbdf56e1297d0887b")
	key, _ := hex.DecodeString(conf.Ue.Key)
	ue.auth.milenage, _ = sec.NewMilenage(key, op, false)
	ue.auth.amf = conf.Ue.Amf
	ue.auth.sqn = conf.Ue.Sqn

	// added supi
	mcc := conf.Ue.Hplmn.Mcc
	mnc := conf.Ue.Hplmn.Mnc

	ue.supi = fmt.Sprintf("imsi-%s%s%s", mcc, mnc, conf.Ue.Msin)

	// added network slice
	ue.Snssai.Sd = conf.Ue.Snssai.Sd
	ue.Snssai.Sst = int(conf.Ue.Snssai.Sst)

	// added Domain Network Name.
	ue.Dnn = conf.Ue.Dnn
	ue.TunnelMode = conf.Ue.TunnelMode

	ue.createSuci(mcc, mnc)

	ue.gnbInboundChannel = gnbInboundChannel
	ue.scenarioChan = scenarioChan

	ue.ExpFile = logFile

	// added initial state for MM(NULL)
	ue.StateMM = MM5G_NULL

	go ue.runService(wg, ueMgrChannel)
	return scenarioChan
}

func (ue *UeContext) CreatePDUSession() (*UEPDUSession, error) {
	pduSessionIndex := -1
	for i, pduSession := range ue.PduSession {
		if pduSession == nil {
			pduSessionIndex = i
			break
		}
	}

	if pduSessionIndex == -1 {
		return nil, errors.New("unable to create an additional PDU Session, we already created the max number of PDU Session")
	}

	pduSession := &UEPDUSession{}
	pduSession.Id = uint8(pduSessionIndex + 1)
	pduSession.Wait = make(chan bool)

	ue.PduSession[pduSessionIndex] = pduSession

	return pduSession, nil
}

func (ue *UeContext) getNasContext() *nas.NasContext {
	if ue.secCtx != nil {
		return ue.secCtx.NasContext()
	}
	return nil
}

func (ue *UeContext) GetUeId() uint8 {
	return ue.id
}

/*
	func (ue *UeContext) GetMsin() string {
		return ue.UeSecurity.Msin
	}

	func (ue *UeContext) GetSupi() string {
		return ue.UeSecurity.Supi
	}
*/
func (ue *UeContext) SetStateMM_DEREGISTERED_INITIATED() {
	ue.StateMM = MM5G_DEREGISTERED_INIT
	ue.scenarioChan <- ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UeContext) SetStateMM_MM5G_SERVICE_REQ_INIT() {
	ue.StateMM = MM5G_SERVICE_REQ_INIT
	ue.scenarioChan <- ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UeContext) SetStateMM_REGISTERED_INITIATED() {
	ue.StateMM = MM5G_REGISTERED_INITIATED
	ue.scenarioChan <- ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UeContext) SetStateMM_REGISTERED() {
	ue.StateMM = MM5G_REGISTERED
	ue.scenarioChan <- ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UeContext) SetStateMM_NULL() {
	ue.StateMM = MM5G_NULL
}

func (ue *UeContext) SetStateMM_DEREGISTERED() {
	ue.StateMM = MM5G_DEREGISTERED
	ue.scenarioChan <- ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UeContext) SetStateMM_IDLE() {
	ue.StateMM = MM5G_IDLE
	ue.scenarioChan <- ScenarioMessage{StateChange: ue.StateMM}
}

func (ue *UeContext) GetStateMM() int {
	return ue.StateMM
}

func (ue *UeContext) SetGnbInboundChannel(gnbInboundChannel chan gnbContext.UEMessage) {
	ue.gnbInboundChannel = gnbInboundChannel
}

func (ue *UeContext) GetGnbInboundChannel() chan gnbContext.UEMessage {
	return ue.gnbInboundChannel
}

func (ue *UeContext) GetDRX() <-chan time.Time {
	if ue.drx == nil {
		return nil
	}
	return ue.drx.C
}

func (ue *UeContext) StopDRX() {
	if ue.drx != nil {
		ue.drx.Stop()
	}
}

func (ue *UeContext) CreateDRX(d time.Duration) {
	ue.drx = time.NewTicker(d)
}

func (ue *UeContext) Lock() {
	ue.lock.Lock()
}

func (ue *UeContext) Unlock() {
	ue.lock.Unlock()
}

func (ue *UeContext) GetPduSession(pduSessionid uint8) (*UEPDUSession, error) {
	if pduSessionid > 16 || ue.PduSession[pduSessionid-1] == nil {
		return nil, errors.New("Unable to find GnbPDUSession ID " + string(pduSessionid))
	}
	return ue.PduSession[pduSessionid-1], nil
}

func (ue *UeContext) GetPduSessions() [16]*gnbContext.GnbPDUSession {
	var pduSessions [16]*gnbContext.GnbPDUSession

	for i, pduSession := range ue.PduSession {
		if pduSession != nil {
			pduSessions[i] = pduSession.GnbPduSession
		}
	}

	return pduSessions
}

func (ue *UeContext) DeletePduSession(pduSessionid uint8) error {
	if pduSessionid > 16 || ue.PduSession[pduSessionid-1] == nil {
		return errors.New("Unable to find GnbPDUSession ID " + string(pduSessionid))
	}
	pduSession := ue.PduSession[pduSessionid-1]
	close(pduSession.Wait)
	stopSignal := pduSession.GetStopSignal()
	if stopSignal != nil {
		stopSignal <- true
	}
	ue.PduSession[pduSessionid-1] = nil
	return nil
}

/*
func (ue *UeContext) GetUeSecurityCapability() *nas.UeSecurityCapability {
	return ue.UeSecurity.UeSecurityCapability
}

func (ue *UeContext) GetMccAndMncInOctets() []byte {
	var res string

	// reverse mcc and mnc
	mcc := reverse(ue.UeSecurity.mcc)
	mnc := reverse(ue.UeSecurity.mnc)

	if len(mnc) == 2 {
		res = fmt.Sprintf("%c%cf%c%c%c", mcc[1], mcc[2], mcc[0], mnc[0], mnc[1])
	} else {
		res = fmt.Sprintf("%c%c%c%c%c%c", mcc[1], mcc[2], mnc[0], mcc[0], mnc[1], mnc[2])
	}

	resu, _ := hex.DecodeString(res)
	return resu
}
*/
// TS 24.501 9.11.3.4.1
// Routing Indicator shall consist of 1 to 4 digits. The coding of this field is the
// responsibility of home network operator but BCD coding shall be used. If a network
// operator decides to assign less than 4 digits to Routing Indicator, the remaining digits
// shall be coded as "1111" to fill the 4 digits coding of Routing Indicator (see NOTE 2). If
// no Routing Indicator is configured in the USIM, the UE shall coxde bits 1 to 4 of octet 8
// of the Routing Indicator as "0000" and the remaining digits as "1111".
/*
func (ue *UeContext) GetRoutingIndicatorInOctets() []byte {
	if len(ue.UeSecurity.RoutingIndicator) == 0 {
		ue.UeSecurity.RoutingIndicator = "0"
	}

	if len(ue.UeSecurity.RoutingIndicator) > 4 {
		log.Fatal("[UE][CONFIG] Routing indicator must be 4 digits maximum, ", ue.UeSecurity.RoutingIndicator, " is invalid")
	}

	routingIndicator := []byte(ue.UeSecurity.RoutingIndicator)
	for len(routingIndicator) < 4 {
		routingIndicator = append(routingIndicator, 'F')
	}

	// Reverse the bytes in group of two
	for i := 1; i < len(routingIndicator); i += 2 {
		tmp := routingIndicator[i-1]
		routingIndicator[i-1] = routingIndicator[i]
		routingIndicator[i] = tmp
	}

	// BCD conversion
	encodedRoutingIndicator, err := hex.DecodeString(string(routingIndicator))
	if err != nil {
		log.Fatal("[UE][CONFIG] Unable to encode routing indicator ", err)
	}

	return encodedRoutingIndicator
}
func (ue *UeContext) getPlmnId() string {
	var plmnId nas.PlmnId
	plmnId.Set(ue.UeSecurity.mcc, ue.UeSecurity.mnc)
	return plmnId.String()
}
*/
func (ue *UeContext) createSuci(mcc, mnc string) {
	var plmnId nas.PlmnId
	plmnId.Set(mcc, mnc)

	suci := new(nas.SupiImsi)
	suci.Parse([]string{plmnId.String(), ue.msin})
	ue.suci = nas.MobileIdentity{
		Id: &nas.Suci{
			Content: suci,
		},
	}
}

/*
	func (ue *UeContext) getAmfRegionId() uint8 {
		return ue.guti.AmfId.GetRegion()
	}

	func (ue *UeContext) getAmfPointer() uint8 {
		return ue.guti.AmfId.GetPointer()
	}

	func (ue *UeContext) getAmfSetId() uint16 {
		return ue.guti.AmfId.GetSet()
	}
*/
func (ue *UeContext) getTMSI5G() (tmsi [4]uint8) {
	if id := ue.guti; id != nil {
		binary.BigEndian.PutUint32(tmsi[:], id.Tmsi)
	}
	return
}

func (ue *UeContext) set5gGuti(guti *nas.MobileIdentity) {
	if guti.GetType() != nas.MobileIdentity5GSType5gGuti {
		//TODO: warn
		return
	}
	ue.guti = guti.Id.(*nas.Guti)
}

/*
func (ue *UeContext) deriveAUTN(autn []byte, ak []uint8) ([]byte, []byte, []byte) {

	sqn := make([]byte, 6)

	// get SQNxorAK
	SQNxorAK := autn[0:6]
	amf := autn[6:8]
	mac_a := autn[8:]

	// get SQN
	for i := 0; i < len(SQNxorAK); i++ {
		sqn[i] = SQNxorAK[i] ^ ak[i]
	}

	// return SQN, amf, mac_a
	return sqn, amf, mac_a
}
*/
/*
func (ue *UeContext) deriveRESstarAndSetKey(authSubs models.AuthenticationSubscription,

	RAND []byte,
	snNmae string,
	AUTN []byte) ([]byte, string) {
	// parameters for authentication challenge.
	mac_a, mac_s := make([]byte, 8), make([]byte, 8)
	CK, IK := make([]byte, 16), make([]byte, 16)
	RES := make([]byte, 8)
	AK, AKstar := make([]byte, 6), make([]byte, 6)

	// Get OPC, K, SQN, AMF from USIM.
	OPC, err := hex.DecodeString(authSubs.EncOpcKey)
	if err != nil {
		log.Fatal("[UE] OPC error: ", err, authSubs.EncOpcKey)
	}
	K, err := hex.DecodeString(authSubs.EncPermanentKey)
	if err != nil {
		log.Fatal("[UE] K error: ", err, authSubs.EncPermanentKey)
	}
	sqnUe, err := hex.DecodeString(authSubs.SequenceNumber.Sqn)
	if err != nil {
		log.Fatal("[UE] sqn error: ", err, authSubs.SequenceNumber.Sqn)
	}
	AMF, err := hex.DecodeString(authSubs.AuthenticationManagementField)
	if err != nil {
		log.Fatal("[UE] AuthenticationManagementField error: ", err, authSubs.AuthenticationManagementField)
	}

	// Generate RES, CK, IK, AK, AKstar
	milenage.F2345(OPC, K, RAND, RES, CK, IK, AK, AKstar)

	// Get SQN, MAC_A, AMF from AUTN
	sqnHn, _, mac_aHn := ue.deriveAUTN(AUTN, AK)

	// Generate MAC_A, MAC_S
	milenage.F1(OPC, K, RAND, sqnHn, AMF, mac_a, mac_s)

	// MAC verification.
	if !reflect.DeepEqual(mac_a, mac_aHn) {
		return nil, "MAC failure"
	}

	// Verification of sequence number freshness.
	if bytes.Compare(sqnUe, sqnHn) > 0 {

		// get AK*
		milenage.F2345(OPC, K, RAND, RES, CK, IK, AK, AKstar)

		// From the standard, AMF(0x0000) should be used in the synch failure.
		amfSynch, _ := hex.DecodeString("0000")

		// get mac_s using sqn ue.
		milenage.F1(OPC, K, RAND, sqnUe, amfSynch, mac_a, mac_s)

		sqnUeXorAK := make([]byte, 6)
		for i := 0; i < len(sqnUe); i++ {
			sqnUeXorAK[i] = sqnUe[i] ^ AKstar[i]
		}

		failureParam := append(sqnUeXorAK, mac_s...)

		return failureParam, "SQN failure"
	}

	// updated sqn value.
	authSubs.SequenceNumber.Sqn = fmt.Sprintf("%x", sqnHn)

	// derive RES*
	key := append(CK, IK...)
	FC := ueauth.FC_FOR_RES_STAR_XRES_STAR_DERIVATION
	P0 := []byte(snNmae)
	P1 := RAND
	P2 := RES

	ue.derivateKamf(key, snNmae, sqnHn, AK)
	kdfVal_for_resStar, err := ueauth.GetKDFValue(key, FC, P0, ueauth.KDFLen(P0), P1, ueauth.KDFLen(P1), P2, ueauth.KDFLen(P2))
	if err != nil {
		log.Fatal("[UE] Error while deriving KDF ", err)
	}
	return kdfVal_for_resStar[len(kdfVal_for_resStar)/2:], "successful"
}

func (ue *UeContext) derivateKamf(key []byte, snName string, SQN, AK []byte) {

	FC := ueauth.FC_FOR_KAUSF_DERIVATION
	P0 := []byte(snName)
	SQNxorAK := make([]byte, 6)
	for i := 0; i < len(SQN); i++ {
		SQNxorAK[i] = SQN[i] ^ AK[i]
	}
	P1 := SQNxorAK
	Kausf, err := ueauth.GetKDFValue(key, FC, P0, ueauth.KDFLen(P0), P1, ueauth.KDFLen(P1))
	if err != nil {
		log.Fatal("[UE] Error while deriving Kausf ", err)
	}
	P0 = []byte(snName)
	Kseaf, err := ueauth.GetKDFValue(Kausf, ueauth.FC_FOR_KSEAF_DERIVATION, P0, ueauth.KDFLen(P0))
	if err != nil {
		log.Fatal("[UE] Error while deriving Kseaf ", err)
	}
	supiRegexp, _ := regexp.Compile("(?:imsi|supi)-([0-9]{5,15})")
	groups := supiRegexp.FindStringSubmatch(ue.UeSecurity.Supi)

	P0 = []byte(groups[1])
	L0 := ueauth.KDFLen(P0)
	P1 = []byte{0x00, 0x00}
	L1 := ueauth.KDFLen(P1)

	ue.UeSecurity.Kamf, err = ueauth.GetKDFValue(Kseaf, ueauth.FC_FOR_KAMF_DERIVATION, P0, L0, P1, L1)
	if err != nil {
		log.Fatal("[UE] Error while deriving Kamf ", err)
	}
}

func (ue *UeContext) DerivateAlgKey() {

	err := auth.AlgorithmKeyDerivation(ue.UeSecurity.CipheringAlg,
		ue.UeSecurity.Kamf,
		&ue.UeSecurity.KnasEnc,
		ue.UeSecurity.IntegrityAlg,
		&ue.UeSecurity.KnasInt)

	if err != nil {
		log.Errorf("[UE] Algorithm key derivation failed  %v", err)
	}
}
*/

func (ue *UeContext) Terminate() {
	ue.SetStateMM_NULL()

	// clean all context of tun interface
	for _, pduSession := range ue.PduSession {
		if pduSession != nil {
			ueTun := pduSession.GetTunInterface()
			ueRule := pduSession.GetTunRule()
			ueRoute := pduSession.GetTunRoute()
			ueVrf := pduSession.GetVrfDevice()

			if ueTun != nil {
				_ = netlink.LinkSetDown(ueTun)
				_ = netlink.LinkDel(ueTun)
			}

			if ueRule != nil {
				_ = netlink.RuleDel(ueRule)
			}

			if ueRoute != nil {
				_ = netlink.RouteDel(ueRoute)
			}

			if ueVrf != nil {
				_ = netlink.LinkSetDown(ueVrf)
				_ = netlink.LinkDel(ueVrf)
			}
		}
	}

	ue.Lock()
	if ue.gnbRx != nil {
		close(ue.gnbRx)
		ue.gnbRx = nil
	}
	if ue.drx != nil {
		ue.drx.Stop()
	}
	ue.Unlock()
	close(ue.scenarioChan)

	log.Info("[UE] UE Terminated")
}

func reverse(s string) string {
	// reverse string.
	var aux string
	for _, valor := range s {
		aux = string(valor) + aux
	}
	return aux
}

func hexCharToByte(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}

	return 0
}
