/**
 * SPDX-License-Identifier: Apache-2.0
 * © Copyright 2023 Hewlett Packard Enterprise Development LP
 * © Copyright 2023-2024 Valentin D'Emmanuele
 */
package ngap

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/lvdund/mssim/internal/control_test_engine/gnb/context"
	"github.com/lvdund/mssim/internal/control_test_engine/gnb/ngap/trigger"

	_ "net"

	"github.com/lvdund/ngap"
	"github.com/lvdund/ngap/aper"
	"github.com/lvdund/ngap/ies"
	"github.com/lvdund/ngap/utils"
	log "github.com/sirupsen/logrus"
	_ "github.com/vishvananda/netlink"
)

// TODO: SEND ERROR INDICATION

func HandlerDownlinkNasTransport(gnb *context.GNBContext, message *ngap.NgapPdu) {

	var ranUeId int64
	var amfUeId int64
	var messageNas []byte

	valueMessage := message.Message.Msg.(*ies.DownlinkNASTransport)

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE NGAP ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE NGAP ID is missing")
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	if valueMessage.NASPDU == nil {
		log.Fatal("[GNB][NGAP] NAS PDU is missing")
	} else {
		messageNas = valueMessage.NASPDU.Value
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot send DownlinkNASTransport message to UE with RANUEID %d as it does not know this UE", ranUeId)
		return
	}

	// send NAS message to UE.
	ue.ReceiveNas(messageNas)
}

func HandlerInitialContextSetupRequest(gnb *context.GNBContext, message *ngap.NgapPdu) {

	var ranUeId int64
	var amfUeId int64
	var messageNas []byte
	var sst []string
	var sd []string
	var mobilityRestrict = "not informed"
	var maskedImeisv string
	var ueSecurityCapabilities *ies.UESecurityCapabilities
	var pDUSessionResourceSetupListCxtReq *ies.PDUSessionResourceSetupListCxtReq
	// var securityKey []byte

	valueMessage := message.Message.Msg.(*ies.InitialContextSetupRequest)

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE NGAP ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] RAN UE NGAP ID is missing")
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	if valueMessage.NASPDU == nil {
		log.Info("[GNB][NGAP] NAS PDU is missing")
	} else {
		messageNas = valueMessage.NASPDU.Value
	}

	// TODO using for create new security context between GNB and UE.
	if valueMessage.SecurityKey == nil {
		log.Fatal("[GNB][NGAP] Security-Key is missing")
	}
	// securityKey = valueMessage.SecurityKey.Value.Bytes

	if valueMessage.GUAMI == nil {
		log.Fatal("[GNB][NGAP] GUAMI is missing")
	}

	if valueMessage.AllowedNSSAI == nil {
		log.Fatal("[GNB][NGAP] Allowed NSSAI is missing")
	}

	valor := len(valueMessage.AllowedNSSAI.Value)
	sst = make([]string, valor)
	sd = make([]string, valor)

	// list S-NSSAI(Single - Network Slice Selection Assistance Information).
	for i, items := range valueMessage.AllowedNSSAI.Value {

		if items.SNSSAI.SST.Value != nil {
			sst[i] = fmt.Sprintf("%x", items.SNSSAI.SST.Value)
		} else {
			sst[i] = "not informed"
		}

		if items.SNSSAI.SD != nil {
			sd[i] = fmt.Sprintf("%x", items.SNSSAI.SD.Value)
		} else {
			sd[i] = "not informed"
		}
	}

	// that field is not mandatory.
	if valueMessage.MobilityRestrictionList == nil {
		log.Info("[GNB][NGAP] Mobility Restriction is missing")
		mobilityRestrict = "not informed"
	} else {
		mobilityRestrict = fmt.Sprintf("%x", valueMessage.MobilityRestrictionList.ServingPLMN.Value)
	}

	// that field is not mandatory.
	// TODO using for mapping UE context
	if valueMessage.MaskedIMEISV == nil {
		log.Info("[GNB][NGAP] Masked IMEISV is missing")
		maskedImeisv = "not informed"
	} else {
		maskedImeisv = fmt.Sprintf("%x", valueMessage.MaskedIMEISV.Value.Bytes)
	}

	// TODO using for create new security context between UE and GNB.
	// TODO algorithms for create new security context between UE and GNB.
	if valueMessage.UESecurityCapabilities == nil {
		log.Fatal("[GNB][NGAP] UE Security Capabilities is missing")
	}
	ueSecurityCapabilities = valueMessage.UESecurityCapabilities

	if valueMessage.PDUSessionResourceSetupListCxtReq == nil {
		log.Warnln("[GNB][NGAP] PDUSessionResourceSetupListCxtReq is missing")
	}
	pDUSessionResourceSetupListCxtReq = valueMessage.PDUSessionResourceSetupListCxtReq

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot setup context for unknown UE	with RANUEID %d", ranUeId)
		return
	}
	// create UE context.
	ue.CreateUeContext(mobilityRestrict, maskedImeisv, sst, sd, ueSecurityCapabilities)

	// show UE context.
	log.Info("[GNB][UE] UE Context was created with successful")
	log.Info("[GNB][UE] UE RAN ID ", ue.GetRanUeId())
	log.Info("[GNB][UE] UE AMF ID ", ue.GetAmfUeId())
	mcc, mnc := ue.GetUeMobility()
	log.Info("[GNB][UE] UE Mobility Restrict --Plmn-- Mcc: ", mcc, " Mnc: ", mnc)
	log.Info("[GNB][UE] UE Masked Imeisv: ", ue.GetUeMaskedImeiSv())
	log.Info("[GNB][UE] Allowed Nssai-- Sst: ", sst, " Sd: ", sd)

	if messageNas != nil {
		ue.ReceiveNas(messageNas)
	}

	if pDUSessionResourceSetupListCxtReq != nil {
		log.Info("[GNB][NGAP] AMF is requesting some PDU Session to be setup during Initial Context Setup")
		for _, pDUSessionResourceSetupItemCtxReq := range pDUSessionResourceSetupListCxtReq.Value {
			pduSessionId := pDUSessionResourceSetupItemCtxReq.PDUSessionID.Value
			sst := fmt.Sprintf("%x", pDUSessionResourceSetupItemCtxReq.SNSSAI.SST.Value)
			sd := "not informed"
			if pDUSessionResourceSetupItemCtxReq.SNSSAI.SD != nil {
				sd = fmt.Sprintf("%x", pDUSessionResourceSetupItemCtxReq.SNSSAI.SD.Value)
			}

			pDUSessionResourceSetupRequestTransferBytes := pDUSessionResourceSetupItemCtxReq.PDUSessionResourceSetupRequestTransfer
			pDUSessionResourceSetupRequestTransfer := &ies.PDUSessionResourceSetupRequestTransfer{}
			if err, _ := pDUSessionResourceSetupRequestTransfer.Decode(*pDUSessionResourceSetupRequestTransferBytes); err != nil {
				log.Error("[GNB] Unable to unmarshall PDUSessionResourceSetupRequestTransfer: ", err)
				continue
			}

			var gtpTunnel *ies.GTPTunnel
			var upfIp string
			var teidUplink aper.OctetString

			if pDUSessionResourceSetupRequestTransfer.ULNGUUPTNLInformation != nil {
				if pDUSessionResourceSetupRequestTransfer.ULNGUUPTNLInformation.GTPTunnel != nil {
					gtpTunnel = pDUSessionResourceSetupRequestTransfer.ULNGUUPTNLInformation.GTPTunnel
					upfIp, _ = utils.IPAddressToString(*gtpTunnel.TransportLayerAddress)
					teidUplink = gtpTunnel.GTPTEID.Value
				}
			}

			if _, err := ue.CreatePduSession(int64(pduSessionId), upfIp, sst, sd, 0, 1, 0, 0, binary.BigEndian.Uint32(teidUplink), gnb.GetUeTeid(ue)); err != nil {
				log.Error("[GNB] ", err)
			}

			if pDUSessionResourceSetupItemCtxReq.NASPDU != nil {
				ue.ReceiveNas(pDUSessionResourceSetupItemCtxReq.NASPDU.Value)
			}
		}

		msg := context.UEMessage{GNBPduSessions: ue.GetPduSessions(), GnbIp: gnb.GetN3GnbIp()}
		ue.ReceiveMessage(&msg)
	}

	// send Initial Context Setup Response.
	log.Info("[GNB][NGAP][AMF] Send Initial Context Setup Response.")
	trigger.SendInitialContextSetupResponse(ue, gnb)
}

func HandlerPduSessionResourceSetupRequest(gnb *context.GNBContext, message *ngap.NgapPdu) {

	var ranUeId int64
	var amfUeId int64
	var pDUSessionResourceSetupList *ies.PDUSessionResourceSetupListSUReq

	// valueMessage := message.InitiatingMessage.Value.PDUSessionResourceSetupRequest
	valueMessage := message.Message.Msg.(*ies.PDUSessionResourceSetupRequest)

	// TODO MORE FIELDS TO CHECK HERE

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] RAN UE ID is missing")
		// TODO SEND ERROR INDICATION
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	if valueMessage.PDUSessionResourceSetupListSUReq == nil {
		log.Fatal("[GNB][NGAP] PDU SESSION RESOURCE SETUP LIST SU REQ is missing")
	} else {
		pDUSessionResourceSetupList = valueMessage.PDUSessionResourceSetupListSUReq
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot setup PDU Session for unknown UE With RANUEID %d", ranUeId)
		return
	}

	var configuredPduSessions []*context.GnbPDUSession
	for _, item := range pDUSessionResourceSetupList.Value {
		var pduSessionId int64
		var ulTeid uint32
		var upfAddress []byte
		var messageNas []byte
		var sst string
		var sd string
		var pduSType uint64
		var qosId int64
		var fiveQi int64
		var priArp int64

		// check PDU Session NAS PDU.
		if item.PDUSessionNASPDU != nil {
			messageNas = item.PDUSessionNASPDU.Value
		} else {
			log.Fatal("[GNB][NGAP] NAS PDU is missing")
		}

		// check pdu session id and nssai information for create a PDU Session.

		// create a PDU session(PDU SESSION ID + NSSAI).
		pduSessionId = int64(item.PDUSessionID.Value)

		if item.SNSSAI.SD != nil {
			sd = fmt.Sprintf("%x", item.SNSSAI.SD.Value)
		} else {
			sd = "not informed"
		}

		if item.SNSSAI.SST.Value != nil {
			sst = fmt.Sprintf("%x", item.SNSSAI.SST.Value)
		} else {
			sst = "not informed"
		}

		if item.PDUSessionResourceSetupRequestTransfer != nil {

			pDUSessionResourceSetupRequestTransfer := &ies.PDUSessionResourceSetupRequestTransfer{}
			if err, _ := pDUSessionResourceSetupRequestTransfer.Decode(*item.PDUSessionResourceSetupRequestTransfer); err == nil {

				ulTeid = binary.BigEndian.Uint32(pDUSessionResourceSetupRequestTransfer.ULNGUUPTNLInformation.GTPTunnel.GTPTEID.Value)
				upfAddress = pDUSessionResourceSetupRequestTransfer.ULNGUUPTNLInformation.GTPTunnel.TransportLayerAddress.Value.Bytes

				for _, itemsQos := range pDUSessionResourceSetupRequestTransfer.QosFlowSetupRequestList.Value {
					qosId = int64(itemsQos.QosFlowIdentifier.Value)
					fiveQi = int64(itemsQos.QosFlowLevelQosParameters.QosCharacteristics.NonDynamic5QI.FiveQI.Value)
					priArp = int64(itemsQos.QosFlowLevelQosParameters.AllocationAndRetentionPriority.PriorityLevelARP.Value)
				}

				pduSType = uint64(pDUSessionResourceSetupRequestTransfer.PDUSessionType.Value)

			} else {
				log.Info("[GNB][NGAP] Error in decode Pdu Session Resource Setup Request Transfer")
			}
		} else {
			log.Fatal("[GNB][NGAP] Error in Pdu Session Resource Setup Request, Pdu Session Resource Setup Request Transfer is missing")
		}

		upfIp := fmt.Sprintf("%d.%d.%d.%d", upfAddress[0], upfAddress[1], upfAddress[2], upfAddress[3])

		// create PDU Session for GNB UE.
		pduSession, err := ue.CreatePduSession(pduSessionId, upfIp, sst, sd, pduSType, qosId, priArp, fiveQi, ulTeid, gnb.GetUeTeid(ue))
		if err != nil {
			log.Error("[GNB][NGAP] Error in Pdu Session Resource Setup Request.")
			log.Error("[GNB][NGAP] ", err)

		}
		configuredPduSessions = append(configuredPduSessions, pduSession)

		log.Info("[GNB][NGAP][UE] PDU Session was created with successful.")
		log.Info("[GNB][NGAP][UE] PDU Session Id: ", pduSession.GetPduSessionId())
		sst, sd = ue.GetSelectedNssai(pduSession.GetPduSessionId())
		log.Info("[GNB][NGAP][UE] NSSAI Selected --- sst: ", sst, " sd: ", sd)
		log.Info("[GNB][NGAP][UE] PDU Session Type: ", pduSession.GetPduType())
		log.Info("[GNB][NGAP][UE] QOS Flow Identifier: ", pduSession.GetQosId())
		log.Info("[GNB][NGAP][UE] Uplink Teid: ", pduSession.GetTeidUplink())
		log.Info("[GNB][NGAP][UE] Downlink Teid: ", pduSession.GetTeidDownlink())
		log.Info("[GNB][NGAP][UE] Non-Dynamic-5QI: ", pduSession.GetFiveQI())
		log.Info("[GNB][NGAP][UE] Priority Level ARP: ", pduSession.GetPriorityARP())
		log.Info("[GNB][NGAP][UE] UPF Address: ", fmt.Sprintf("%d.%d.%d.%d", upfAddress[0], upfAddress[1], upfAddress[2], upfAddress[3]), " :2152")

		// send NAS message to UE.
		ue.ReceiveNas(messageNas)

		var pduSessions [16]*context.GnbPDUSession
		pduSessions[0] = pduSession
		msg := context.UEMessage{GnbIp: gnb.GetN3GnbIp(), GNBPduSessions: pduSessions}

		ue.ReceiveMessage(&msg)
	}

	// send PDU Session Resource Setup Response.
	trigger.SendPduSessionResourceSetupResponse(configuredPduSessions, ue, gnb)
}

func HandlerPduSessionReleaseCommand(gnb *context.GNBContext, message *ngap.NgapPdu) {
	valueMessage := message.Message.Msg.(*ies.PDUSessionResourceReleaseCommand)

	var amfUeId int64
	var ranUeId int64
	var messageNas aper.OctetString
	var pduSessionIds []ies.PDUSessionID

	// TODO MORE FIELDS TO CHECK HERE

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] RAN UE ID is missing")
		// TODO SEND ERROR INDICATION
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	if valueMessage.NASPDU == nil {
		log.Info("[GNB][NGAP] NAS PDU is missing")
		// TODO SEND ERROR INDICATION
	} else {
		messageNas = valueMessage.NASPDU.Value
	}

	if valueMessage.PDUSessionResourceToReleaseListRelCmd == nil {
		log.Fatal("[GNB][NGAP] PDU SESSION RESOURCE SETUP LIST SU REQ is missing")
	} else {
		for _, pDUSessionRessourceToReleaseItemRelCmd := range valueMessage.PDUSessionResourceToReleaseListRelCmd.Value {
			pduSessionIds = append(pduSessionIds, *pDUSessionRessourceToReleaseItemRelCmd.PDUSessionID)
		}
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot release PDU Session for unknown UE With RANUEID %d", ranUeId)
		return
	}

	for _, pduSessionId := range pduSessionIds {
		pduSession, err := ue.GetPduSession(int64(pduSessionId.Value))
		if pduSession == nil || err != nil {
			log.Error("[GNB][NGAP] Unable to delete PDU Session ", pduSessionId.Value, " from UE as the PDU Session was not found. Ignoring.")
			continue
		}
		ue.DeletePduSession(int64(pduSessionId.Value))
		log.Info("[GNB][NGAP] Successfully deleted PDU Session ", pduSessionId.Value, " from UE Context")
	}

	trigger.SendPduSessionReleaseResponse(pduSessionIds, ue)

	ue.ReceiveNas(messageNas)
}

func HandlerNgSetupResponse(amf *context.GNBAmf, gnb *context.GNBContext, message *ngap.NgapPdu) {

	err := false
	var plmn string

	// check information about AMF and add in AMF context.
	valueMessage := message.Message.Msg.(*ies.NGSetupResponse)

	if valueMessage.AMFName == nil {
		// TODO error indication. This field is mandatory critically reject
		log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,AMF Name is missing")
		log.Info("[GNB][NGAP] AMF is inactive")
		err = true
	} else {
		amfName := valueMessage.AMFName.Value
		amf.SetAmfName(string(amfName))
	}

	if valueMessage.ServedGUAMIList.Value == nil {
		// TODO error indication. This field is mandatory critically reject
		log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Serverd Guami list is missing")
		log.Info("[GNB][NGAP] AMF is inactive")
		err = true
	}
	for _, items := range valueMessage.ServedGUAMIList.Value {
		if items.GUAMI.AMFRegionID.Value.Bytes == nil {
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Served Guami list is inappropriate")
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, AMFRegionId is missing")
			log.Info("[GNB][NGAP] AMF is inactive")
			err = true
		}
		if items.GUAMI.AMFPointer.Value.Bytes == nil {
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Served Guami list is inappropriate")
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, AMFPointer is missing")
			log.Info("[GNB][NGAP] AMF is inactive")
			err = true
		}
		if items.GUAMI.AMFSetID.Value.Bytes == nil {
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE,Served Guami list is inappropriate")
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, AMFSetId is missing")
			log.Info("[GNB][NGAP] AMF is inactive")
			err = true
		}
	}

	if valueMessage.RelativeAMFCapacity != nil {
		amfCapacity := valueMessage.RelativeAMFCapacity.Value
		amf.SetAmfCapacity(int64(amfCapacity))
	}

	if valueMessage.PLMNSupportList == nil {
		log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, PLMN Support list is missing")
		err = true
	}

	for _, items := range valueMessage.PLMNSupportList.Value {

		plmn = fmt.Sprintf("%x", items.PLMNIdentity.Value)
		amf.AddedPlmn(plmn)

		if items.SliceSupportList.Value == nil {
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, PLMN Support list is inappropriate")
			log.Info("[GNB][NGAP] Error in NG SETUP RESPONSE, Slice Support list is missing")
			err = true
		}

		for _, slice := range items.SliceSupportList.Value {

			var sd string
			var sst string

			if slice.SNSSAI.SST.Value != nil {
				sst = fmt.Sprintf("%x", slice.SNSSAI.SST.Value)
			} else {
				sst = "was not informed"
			}

			if slice.SNSSAI.SD != nil {
				sd = fmt.Sprintf("%x", slice.SNSSAI.SD.Value)
			} else {
				sd = "was not informed"
			}

			// update amf slice supported
			amf.AddedSlice(sst, sd)
		}
	}

	if err {
		log.Fatal("[GNB][AMF] AMF is inactive")
		amf.SetStateInactive()
	} else {
		amf.SetStateActive()
		log.Info("[GNB][AMF] AMF Name: ", amf.GetAmfName())
		log.Info("[GNB][AMF] State of AMF: Active")
		log.Info("[GNB][AMF] Capacity of AMF: ", amf.GetAmfCapacity())
		for i := 0; i < amf.GetLenPlmns(); i++ {
			mcc, mnc := amf.GetPlmnSupport(i)
			log.Info("[GNB][AMF] PLMNs Identities Supported by AMF -- mcc: ", mcc, " mnc:", mnc)
		}
		for i := 0; i < amf.GetLenSlice(); i++ {
			sst, sd := amf.GetSliceSupport(i)
			log.Info("[GNB][AMF] List of AMF slices Supported by AMF -- sst:", sst, " sd:", sd)
		}
	}

}

func HandlerNgSetupFailure(amf *context.GNBAmf, gnb *context.GNBContext, message *ngap.NgapPdu) {

	// check information about AMF and add in AMF context.
	valueMessage := message.Message.Msg.(*ies.NGSetupFailure)

	if valueMessage.Cause != nil {
		log.Error("[GNB][NGAP] Received failure from AMF: ", causeToString(valueMessage.Cause))
	}

	// redundant but useful for information about code.
	amf.SetStateInactive()

	log.Info("[GNB][NGAP] AMF is inactive")
}

func HandlerUeContextReleaseCommand(gnb *context.GNBContext, message *ngap.NgapPdu) {

	valueMessage := message.Message.Msg.(*ies.UEContextReleaseCommand)

	var cause *ies.Cause
	var ue_id *ies.RANUENGAPID

	if valueMessage.UENGAPIDs != nil {
		ue_id = valueMessage.UENGAPIDs.UENGAPIDpair.RANUENGAPID
	}
	if valueMessage.Cause != nil {
		cause = valueMessage.Cause
	}

	ue, err := gnb.GetGnbUe(int64(ue_id.Value))
	if err != nil {
		log.Error("[GNB][AMF] AMF is trying to free the context of an unknown UE")
		return
	}
	gnb.DeleteGnBUe(ue)

	// Send UEContextReleaseComplete
	trigger.SendUeContextReleaseComplete(ue)

	log.Info("[GNB][NGAP] Releasing UE Context, cause: ", causeToString(cause))
}

func HandlerAmfConfigurationUpdate(amf *context.GNBAmf, gnb *context.GNBContext, message *ngap.NgapPdu) {
	log.Debugf("Before Update:")

	amfPool := gnb.GetAmfPool()
	amfPool.Range(func(k, v any) bool {
		oldAmf, ok := v.(*context.GNBAmf)
		if ok {
			tnla := oldAmf.GetTNLA()
			log.Debugf("[AMF Name: %5s], IP: %10s, AMFCapacity: %3d, TNLA Weight Factor: %2d, TNLA Usage: %2d\n",
				oldAmf.GetAmfName(), oldAmf.GetAmfIp(), oldAmf.GetAmfCapacity(), tnla.GetWeightFactor(), tnla.GetUsage())
		}
		return true
	})

	var amfName string
	var amfCapacity int64
	var amfRegionId, amfSetId, amfPointer aper.BitString

	valueMessage := message.Message.Msg.(*ies.AMFConfigurationUpdate)

	if valueMessage.AMFName != nil {
		amfName = string(valueMessage.AMFName.Value)
	}

	if valueMessage.ServedGUAMIList != nil {
		for _, servedGuamiItem := range valueMessage.ServedGUAMIList.Value {
			amfRegionId = servedGuamiItem.GUAMI.AMFRegionID.Value
			amfSetId = servedGuamiItem.GUAMI.AMFSetID.Value
			amfPointer = servedGuamiItem.GUAMI.AMFPointer.Value
		}
	}

	if valueMessage.RelativeAMFCapacity != nil {
		amfCapacity = int64(valueMessage.RelativeAMFCapacity.Value)
	}

	if valueMessage.AMFTNLAssociationToAddList != nil {
		toAddList := valueMessage.AMFTNLAssociationToAddList
		for _, toAddItem := range toAddList.Value {
			bitLen := toAddItem.AMFTNLAssociationAddress.EndpointIPAddress.Value.NumBits
			var ipv4String string
			if bitLen == 32 || bitLen == 160 { // IPv4 or IPv4+IPv6
				ipv4String, _ = utils.IPAddressToString(*toAddItem.AMFTNLAssociationAddress.EndpointIPAddress)
			}

			amfPool := gnb.GetAmfPool()
			amfExisted := false
			amfPool.Range(func(key, value any) bool {
				gnbAmf, ok := value.(*context.GNBAmf)
				if !ok {
					return true
				}
				if gnbAmf.GetAmfIp() == ipv4String {
					log.Info("[GNB] SCTP/NGAP service exists")
					amfExisted = true
					return false
				}
				return true
			})
			if amfExisted {
				continue
			}

			port := 38412 // default sctp port
			newAmf := gnb.NewGnBAmf(ipv4String, port)
			newAmf.SetAmfName(amfName)
			newAmf.SetAmfCapacity(amfCapacity)
			newAmf.SetRegionId(amfRegionId)
			newAmf.SetSetId(amfSetId)
			newAmf.SetPointer(amfPointer)
			newAmf.SetTNLAUsage(toAddItem.TNLAssociationUsage.Value)
			newAmf.SetTNLAWeight(int64(toAddItem.TNLAddressWeightFactor.Value))

			// start communication with AMF(SCTP).
			if err := InitConn(newAmf, gnb); err != nil {
				log.Fatal("Error in", err)
			} else {
				log.Info("[GNB] SCTP/NGAP service is running")
				// wg.Add(1)
			}

			trigger.SendNgSetupRequest(gnb, newAmf)

		}
	}

	if valueMessage.AMFTNLAssociationToRemoveList != nil {
		toRemoveList := valueMessage.AMFTNLAssociationToRemoveList
		for _, toRemoveItem := range toRemoveList.Value {
			bitLen := toRemoveItem.AMFTNLAssociationAddress.EndpointIPAddress.Value.NumBits
			var ipv4String string
			if bitLen == 32 || bitLen == 160 { // IPv4 or IPv4+IPv6
				ipv4String, _ = utils.IPAddressToString(*toRemoveItem.AMFTNLAssociationAddress.EndpointIPAddress)
			}
			port := 38412 // default sctp port
			amfPool := gnb.GetAmfPool()
			amfPool.Range(func(k, v any) bool {
				oldAmf, ok := v.(*context.GNBAmf)
				if ok && oldAmf.GetAmfIp() == ipv4String && oldAmf.GetAmfPort() == port {
					log.Info("[GNB][AMF] Remove AMF:", amf.GetAmfName(), " IP:", amf.GetAmfIp())
					tnla := amf.GetTNLA()
					tnla.Release() // Close SCTP Conntection
					amfPool.Delete(k)
					return false
				}
				return true
			})
		}
	}

	if valueMessage.AMFTNLAssociationToUpdateList != nil {
		toUpdateList := valueMessage.AMFTNLAssociationToUpdateList
		for _, toUpdateItem := range toUpdateList.Value {
			bitLen := toUpdateItem.AMFTNLAssociationAddress.EndpointIPAddress.Value.NumBits
			var ipv4String string
			if bitLen == 32 || bitLen == 160 {
				ipv4String, _ = utils.IPAddressToString(*toUpdateItem.AMFTNLAssociationAddress.EndpointIPAddress)
			}
			port := 38412 // default sctp port
			amfPool := gnb.GetAmfPool()
			amfPool.Range(func(k, v any) bool {
				oldAmf, ok := v.(*context.GNBAmf)
				if ok && oldAmf.GetAmfIp() == ipv4String && oldAmf.GetAmfPort() == port {
					oldAmf.SetAmfName(amfName)
					oldAmf.SetAmfCapacity(amfCapacity)
					oldAmf.SetRegionId(amfRegionId)
					oldAmf.SetSetId(amfSetId)
					oldAmf.SetPointer(amfPointer)

					oldAmf.SetTNLAUsage(toUpdateItem.TNLAssociationUsage.Value)
					oldAmf.SetTNLAWeight(int64(toUpdateItem.TNLAddressWeightFactor.Value))
					return false
				}
				return true
			})
		}
	}

	log.Debugf("After Update:")
	amfPool = gnb.GetAmfPool()
	amfPool.Range(func(k, v any) bool {
		oldAmf, ok := v.(*context.GNBAmf)
		if ok {
			tnla := oldAmf.GetTNLA()
			log.Debugf("[AMF Name: %5s], IP: %10s, AMFCapacity: %3d, TNLA Weight Factor: %2d, TNLA Usage: %2d\n",
				oldAmf.GetAmfName(), oldAmf.GetAmfIp(), oldAmf.GetAmfCapacity(), tnla.GetWeightFactor(), tnla.GetUsage())
		}
		return true
	})

	trigger.SendAmfConfigurationUpdateAcknowledge(amf)
}

func HandlerAmfStatusIndication(amf *context.GNBAmf, gnb *context.GNBContext, message *ngap.NgapPdu) {
	valueMessage := message.Message.Msg.(*ies.AMFStatusIndication)

	if valueMessage.UnavailableGUAMIList != nil {
		for _, unavailableGuamiItem := range valueMessage.UnavailableGUAMIList.Value {
			octetStr := unavailableGuamiItem.GUAMI.PLMNIdentity.Value
			hexStr := fmt.Sprintf("%02x%02x%02x", octetStr[0], octetStr[1], octetStr[2])
			var unavailableMcc, unavailableMnc string
			unavailableMcc = string(hexStr[1]) + string(hexStr[0]) + string(hexStr[3])
			unavailableMnc = string(hexStr[5]) + string(hexStr[4])
			if hexStr[2] != 'f' {
				unavailableMnc = string(hexStr[2]) + string(hexStr[5]) + string(hexStr[4])
			}

			amfPool := gnb.GetAmfPool()

			// select backup AMF
			var backupAmf *context.GNBAmf
			amfPool.Range(func(k, v any) bool {
				amf, ok := v.(*context.GNBAmf)
				if !ok {
					return true
				}
				if unavailableGuamiItem.BackupAMFName != nil &&
					amf.GetAmfName() == string(unavailableGuamiItem.BackupAMFName.Value) {
					backupAmf = amf
					return false
				}

				return true
			})

			if backupAmf == nil {
				return
			}

			amfPool.Range(func(k, v any) bool {
				oldAmf, ok := v.(*context.GNBAmf)
				if !ok {
					return true
				}
				for j := 0; j < oldAmf.GetLenPlmns(); j++ {
					oldAmfSupportMcc, oldAmfSupportMnc := oldAmf.GetPlmnSupport(j)

					if oldAmfSupportMcc == unavailableMcc && oldAmfSupportMnc == unavailableMnc &&
						reflect.DeepEqual(oldAmf.GetRegionId(), unavailableGuamiItem.GUAMI.AMFRegionID.Value) &&
						reflect.DeepEqual(oldAmf.GetSetId(), unavailableGuamiItem.GUAMI.AMFSetID.Value) &&
						reflect.DeepEqual(oldAmf.GetPointer(), unavailableGuamiItem.GUAMI.AMFPointer.Value) {

						log.Info("[GNB][AMF] Remove AMF: [",
							"Id: ", oldAmf.GetAmfId(),
							"Name: ", oldAmf.GetAmfName(),
							"Ipv4: ", oldAmf.GetAmfIp(),
							"]",
						)

						tnla := oldAmf.GetTNLA()

						// NGAP UE-TNLA Rebinding
						uePool := gnb.GetUePool()
						uePool.Range(func(k, v any) bool {
							ue, ok := v.(*context.GNBUe)
							if !ok {
								return true
							}

							if ue.GetAmfId() == oldAmf.GetAmfId() {
								// set amfId and SCTP association for UE.
								ue.SetAmfId(backupAmf.GetAmfId())
								ue.SetSCTP(backupAmf.GetSCTPConn())
							}

							return true
						})

						prUePool := gnb.GetPrUePool()
						prUePool.Range(func(k, v any) bool {
							ue, ok := v.(*context.GNBUe)
							if !ok {
								return true
							}

							if ue.GetAmfId() == oldAmf.GetAmfId() {
								// set amfId and SCTP association for UE.
								ue.SetAmfId(backupAmf.GetAmfId())
								ue.SetSCTP(backupAmf.GetSCTPConn())
							}

							return true
						})

						tnla.Release()
						amfPool.Delete(k)

						return true
					}
				}
				return true
			})
		}
	}
}

func HandlerPathSwitchRequestAcknowledge(gnb *context.GNBContext, message *ngap.NgapPdu) {
	var pduSessionResourceSwitchedList *ies.PDUSessionResourceSwitchedList
	valueMessage := message.Message.Msg.(*ies.PathSwitchRequestAcknowledge)

	var amfUeId, ranUeId int64

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] RAN UE ID is missing")
		// TODO SEND ERROR INDICATION
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	pduSessionResourceSwitchedList = valueMessage.PDUSessionResourceSwitchedList
	if pduSessionResourceSwitchedList == nil {
		log.Fatal("[GNB][NGAP] PduSessionResourceSwitchedList is missing")
		// TODO SEND ERROR INDICATION
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot Xn Handover unknown UE With RANUEID %d", ranUeId)
		return
	}

	if pduSessionResourceSwitchedList == nil || len(pduSessionResourceSwitchedList.Value) == 0 {
		log.Warn("[GNB] No PDU Sessions to be switched")
		return
	}

	for _, pduSessionResourceSwitchedItem := range pduSessionResourceSwitchedList.Value {
		pduSessionId := pduSessionResourceSwitchedItem.PDUSessionID.Value
		pduSession, err := ue.GetPduSession(int64(pduSessionId))
		if err != nil {
			log.Error("[GNB] Trying to path switch an unknown PDU Session ID ", pduSessionId, ": ", err)
			continue
		}

		pathSwitchRequestAcknowledgeTransferBytes := pduSessionResourceSwitchedItem.PathSwitchRequestAcknowledgeTransfer
		pathSwitchRequestAcknowledgeTransfer := &ies.PathSwitchRequestAcknowledgeTransfer{}
		// err = aper.UnmarshalWithParams(pathSwitchRequestAcknowledgeTransferBytes, pathSwitchRequestAcknowledgeTransfer, "valueExt")
		err = pathSwitchRequestAcknowledgeTransfer.Decode(aper.NewReader(bytes.NewBuffer(*pathSwitchRequestAcknowledgeTransferBytes)))
		if err != nil {
			log.Error("[GNB] Unable to unmarshall PathSwitchRequestAcknowledgeTransfer: ", err)
			continue
		}

		if pathSwitchRequestAcknowledgeTransfer.ULNGUUPTNLInformation != nil {
			gtpTunnel := pathSwitchRequestAcknowledgeTransfer.ULNGUUPTNLInformation.GTPTunnel
			upfIpv4, _ := utils.IPAddressToString(*gtpTunnel.TransportLayerAddress)
			teidUplink := gtpTunnel.GTPTEID.Value

			// Set new Teid Uplink received in PathSwitchRequestAcknowledge
			pduSession.SetTeidUplink(binary.BigEndian.Uint32(teidUplink))
			pduSession.SetUpfIp(upfIpv4)
		}
		var pduSessions [16]*context.GnbPDUSession
		pduSessions[0] = pduSession

		msg := context.UEMessage{GNBPduSessions: pduSessions, GnbIp: gnb.GetN3GnbIp()}

		ue.ReceiveMessage(&msg)
	}

	log.Info("[GNB] Handover completed successfully for UE ", ue.GetRanUeId())
}

func HandlerHandoverRequest(amf *context.GNBAmf, gnb *context.GNBContext, message *ngap.NgapPdu) {
	var ueSecurityCapabilities *ies.UESecurityCapabilities
	var sst []string
	var sd []string
	var maskedImeisv string
	var sourceToTargetContainer *ies.SourceToTargetTransparentContainer
	var pDUSessionResourceSetupListHOReq *ies.PDUSessionResourceSetupListHOReq
	var amfUeId int64

	valueMessage := message.Message.Msg.(*ies.HandoverRequest)

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.AllowedNSSAI == nil {
		log.Fatal("[GNB][NGAP] Allowed NSSAI is missing")
	} else {
		valor := len(valueMessage.AllowedNSSAI.Value)
		sst = make([]string, valor)
		sd = make([]string, valor)

		// list S-NSSAI(Single - Network Slice Selection Assistance Information).
		for i, items := range valueMessage.AllowedNSSAI.Value {

			if items.SNSSAI.SST.Value != nil {
				sst[i] = fmt.Sprintf("%x", items.SNSSAI.SST.Value)
			} else {
				sst[i] = "not informed"
			}

			if items.SNSSAI.SD != nil {
				sd[i] = fmt.Sprintf("%x", items.SNSSAI.SD.Value)
			} else {
				sd[i] = "not informed"
			}
		}
	}

	// that field is not mandatory.
	// TODO using for mapping UE context
	if valueMessage.MaskedIMEISV == nil {
		log.Info("[GNB][NGAP] Masked IMEISV is missing")
		maskedImeisv = "not informed"
	} else {
		maskedImeisv = fmt.Sprintf("%x", valueMessage.MaskedIMEISV.Value.Bytes)
	}

	sourceToTargetContainer = valueMessage.SourceToTargetTransparentContainer
	if sourceToTargetContainer == nil {
		log.Fatal("[GNB][NGAP] sourceToTargetContainer is missing")
		// TODO SEND ERROR INDICATION
	}

	pDUSessionResourceSetupListHOReq = valueMessage.PDUSessionResourceSetupListHOReq
	if pDUSessionResourceSetupListHOReq == nil {
		log.Fatal("[GNB][NGAP] pDUSessionResourceSetupListHOReq is missing")
		// TODO SEND ERROR INDICATION
	}

	if valueMessage.UESecurityCapabilities == nil {
		log.Fatal("[GNB][NGAP] UE Security Capabilities is missing")
	} else {
		ueSecurityCapabilities = valueMessage.UESecurityCapabilities
	}

	if sourceToTargetContainer == nil {
		log.Error("[GNB] HandoverRequest message from AMF is missing mandatory SourceToTargetTransparentContainer")
		return
	}

	sourceToTargetContainerBytes := sourceToTargetContainer.Value
	sourceToTargetContainerNgap := &ies.SourceNGRANNodeToTargetNGRANNodeTransparentContainer{}
	// err := aper.UnmarshalWithParams(sourceToTargetContainerBytes, sourceToTargetContainerNgap, "valueExt")
	err := sourceToTargetContainer.Decode(aper.NewReader(bytes.NewBuffer(sourceToTargetContainerBytes)))
	if err != nil {
		log.Error("[GNB] Unable to unmarshall SourceToTargetTransparentContainer: ", err)
		return
	}
	if sourceToTargetContainerNgap.IndexToRFSP == nil {
		log.Error("[GNB] SourceToTargetTransparentContainer from source gNodeB is missing IndexToRFSP")
		return
	}
	prUeId := sourceToTargetContainerNgap.IndexToRFSP.Value

	ue, err := gnb.NewGnBUe(nil, nil, int64(prUeId), nil)
	if ue == nil || err != nil {
		log.Fatalf("[GNB] HandoverFailure: %s", err)
	}
	ue.SetAmfUeId(amfUeId)

	ue.CreateUeContext("not informed", maskedImeisv, sst, sd, ueSecurityCapabilities)

	for _, pDUSessionResourceSetupItemHOReq := range pDUSessionResourceSetupListHOReq.Value {
		pduSessionId := pDUSessionResourceSetupItemHOReq.PDUSessionID.Value
		sst := fmt.Sprintf("%x", pDUSessionResourceSetupItemHOReq.SNSSAI.SST.Value)
		sd := "not informed"
		if pDUSessionResourceSetupItemHOReq.SNSSAI.SD != nil {
			sd = fmt.Sprintf("%x", pDUSessionResourceSetupItemHOReq.SNSSAI.SD.Value)
		}

		handOverRequestTransferBytes := pDUSessionResourceSetupItemHOReq.HandoverRequestTransfer
		handOverRequestTransfer := &ies.PDUSessionResourceSetupRequestTransfer{}
		if err, _ := handOverRequestTransfer.Decode(*handOverRequestTransferBytes); err != nil {
			log.Error("[GNB] Unable to unmarshall HandOverRequestTransfer: ", err)
			continue
		}

		var gtpTunnel *ies.GTPTunnel
		var upfIp string
		var teidUplink aper.OctetString

		if handOverRequestTransfer.ULNGUUPTNLInformation != nil {
			uLNGUUPTNLInformation := handOverRequestTransfer.ULNGUUPTNLInformation

			gtpTunnel = uLNGUUPTNLInformation.GTPTunnel
			upfIp, _ = utils.IPAddressToString(*gtpTunnel.TransportLayerAddress)
			teidUplink = gtpTunnel.GTPTEID.Value
		}

		_, err = ue.CreatePduSession(int64(pduSessionId), upfIp, sst, sd, 0, 1, 0, 0, binary.BigEndian.Uint32(teidUplink), gnb.GetUeTeid(ue))
		if err != nil {
			log.Error("[GNB] ", err)
		}
	}

	trigger.SendHandoverRequestAcknowledge(gnb, ue)
}

func HandlerHandoverCommand(amf *context.GNBAmf, gnb *context.GNBContext, message *ngap.NgapPdu) {
	valueMessage := message.Message.Msg.(*ies.HandoverCommand)

	var amfUeId, ranUeId int64

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] RAN UE ID is missing")
		// TODO SEND ERROR INDICATION
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	ue := getUeFromContext(gnb, ranUeId, amfUeId)
	if ue == nil {
		log.Errorf("[GNB][NGAP] Cannot NGAP  Handover unknown UE With RANUEID %d", ranUeId)
		return
	}
	newGnb := ue.GetHandoverGnodeB()
	if newGnb == nil {
		log.Error("[GNB] AMF is sending a Handover Command for an UE we did not send a Handover Required message")
		// TODO SEND ERROR INDICATION
		return
	}

	newGnbRx := make(chan context.UEMessage, 1)
	newGnbTx := make(chan context.UEMessage, 1)
	newGnb.GetInboundChannel() <- context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, PrUeId: ue.GetPrUeId(), IsHandover: true}

	msg := context.UEMessage{GNBRx: newGnbRx, GNBTx: newGnbTx, GNBInboundChannel: newGnb.GetInboundChannel()}

	ue.ReceiveMessage(&msg)
}

func HandlerPaging(gnb *context.GNBContext, message *ngap.NgapPdu) {

	valueMessage := message.Message.Msg.(*ies.Paging)

	var uEPagingIdentity *ies.UEPagingIdentity
	var tAIListForPaging *ies.TAIListForPaging

	if valueMessage.UEPagingIdentity == nil {
		log.Fatal("[GNB][NGAP] UE Paging Identity is missing")
	} else {
		uEPagingIdentity = valueMessage.UEPagingIdentity
	}

	if valueMessage.TAIListForPaging == nil {
		log.Fatal("[GNB][NGAP] TAI List For Paging is missing")
	} else {
		tAIListForPaging = valueMessage.TAIListForPaging
	}

	_ = tAIListForPaging

	gnb.AddPagedUE(uEPagingIdentity.FiveGSTMSI)

	log.Info("[GNB][AMF] Paging UE")
}

func HandlerErrorIndication(gnb *context.GNBContext, message *ngap.NgapPdu) {

	valueMessage := message.Message.Msg.(*ies.ErrorIndication)

	var amfUeId, ranUeId int64

	if valueMessage.AMFUENGAPID == nil {
		log.Fatal("[GNB][NGAP] AMF UE ID is missing")
	} else {
		amfUeId = int64(valueMessage.AMFUENGAPID.Value)
	}

	if valueMessage.RANUENGAPID == nil {
		log.Fatal("[GNB][NGAP] RAN UE ID is missing")
		// TODO SEND ERROR INDICATION
	} else {
		ranUeId = int64(valueMessage.RANUENGAPID.Value)
	}

	log.Warn("[GNB][AMF] Received an Error Indication for UE with AMF UE ID: ", amfUeId, " RAN UE ID: ", ranUeId)
}

func getUeFromContext(gnb *context.GNBContext, ranUeId int64, amfUeId int64) *context.GNBUe {
	// check RanUeId and get UE.
	ue, err := gnb.GetGnbUe(ranUeId)
	if err != nil || ue == nil {
		log.Error("[GNB][NGAP] RAN UE NGAP ID is incorrect, found: ", ranUeId)
		return nil
		// TODO SEND ERROR INDICATION
	}

	ue.SetAmfUeId(amfUeId)

	return ue
}

func causeToString(cause *ies.Cause) string {
	if cause != nil {
		switch cause.Choice {
		case uint64(ies.CausePresentRadioNetwork):
			return "radioNetwork: " + causeRadioNetworkToString(cause.RadioNetwork)
		case uint64(ies.CausePresentTransport):
			return "transport: " + causeTransportToString(cause.Transport)
		case uint64(ies.CausePresentNas):
			return "nas: " + causeNasToString(cause.Nas)
		case uint64(ies.CausePresentProtocol):
			return "protocol: " + causeProtocolToString(cause.Protocol)
		case uint64(ies.CausePresentMisc):
			return "misc: " + causeMiscToString(cause.Misc)
		}
	}
	return "Cause not found"
}

func causeRadioNetworkToString(network *ies.CauseRadioNetwork) string {
	switch network.Value {
	case ies.CauseRadioNetworkUnspecified:
		return "Unspecified cause for radio network"
	case ies.CauseRadioNetworkTxnrelocoverallexpiry:
		return "Transfer the overall timeout of radio resources during handover"
	case ies.CauseRadioNetworkSuccessfulhandover:
		return "Successful handover"
	case ies.CauseRadioNetworkReleaseduetongrangeneratedreason:
		return "Release due to NG-RAN generated reason"
	case ies.CauseRadioNetworkReleasedueto5Gcgeneratedreason:
		return "Release due to 5GC generated reason"
	case ies.CauseRadioNetworkHandovercancelled:
		return "Handover cancelled"
	case ies.CauseRadioNetworkPartialhandover:
		return "Partial handover"
	case ies.CauseRadioNetworkHofailureintarget5Gcngrannodeortargetsystem:
		return "Handover failure in target 5GC NG-RAN node or target system"
	case ies.CauseRadioNetworkHotargetnotallowed:
		return "Handover target not allowed"
	case ies.CauseRadioNetworkTngrelocoverallexpiry:
		return "Transfer the overall timeout of radio resources during target NG-RAN relocation"
	case ies.CauseRadioNetworkTngrelocprepexpiry:
		return "Transfer the preparation timeout of radio resources during target NG-RAN relocation"
	case ies.CauseRadioNetworkCellnotavailable:
		return "Cell not available"
	case ies.CauseRadioNetworkUnknowntargetid:
		return "Unknown target ID"
	case ies.CauseRadioNetworkNoradioresourcesavailableintargetcell:
		return "No radio resources available in the target cell"
	case ies.CauseRadioNetworkUnknownlocaluengapid:
		return "Unknown local UE NGAP ID"
	case ies.CauseRadioNetworkInconsistentremoteuengapid:
		return "Inconsistent remote UE NGAP ID"
	case ies.CauseRadioNetworkHandoverdesirableforradioreason:
		return "Handover desirable for radio reason"
	case ies.CauseRadioNetworkTimecriticalhandover:
		return "Time-critical handover"
	case ies.CauseRadioNetworkResourceoptimisationhandover:
		return "Resource optimization handover"
	case ies.CauseRadioNetworkReduceloadinservingcell:
		return "Reduce load in serving cell"
	case ies.CauseRadioNetworkUserinactivity:
		return "User inactivity"
	case ies.CauseRadioNetworkRadioconnectionwithuelost:
		return "Radio connection with UE lost"
	case ies.CauseRadioNetworkRadioresourcesnotavailable:
		return "Radio resources not available"
	case ies.CauseRadioNetworkInvalidqoscombination:
		return "Invalid QoS combination"
	case ies.CauseRadioNetworkFailureinradiointerfaceprocedure:
		return "Failure in radio interface procedure"
	case ies.CauseRadioNetworkInteractionwithotherprocedure:
		return "Interaction with other procedure"
	case ies.CauseRadioNetworkUnknownpdusessionid:
		return "Unknown PDU session ID"
	case ies.CauseRadioNetworkUnkownqosflowid:
		return "Unknown QoS flow ID"
	case ies.CauseRadioNetworkMultiplepdusessionidinstances:
		return "Multiple PDU session ID instances"
	case ies.CauseRadioNetworkMultipleqosflowidinstances:
		return "Multiple QoS flow ID instances"
	case ies.CauseRadioNetworkEncryptionandorintegrityprotectionalgorithmsnotsupported:
		return "Encryption and/or integrity protection algorithms not supported"
	case ies.CauseRadioNetworkNgintrasystemhandovertriggered:
		return "NG intra-system handover triggered"
	case ies.CauseRadioNetworkNgintersystemhandovertriggered:
		return "NG inter-system handover triggered"
	case ies.CauseRadioNetworkXnhandovertriggered:
		return "Xn handover triggered"
	case ies.CauseRadioNetworkNotsupported5Qivalue:
		return "Not supported 5QI value"
	case ies.CauseRadioNetworkUecontexttransfer:
		return "UE context transfer"
	case ies.CauseRadioNetworkImsvoiceepsfallbackorratfallbacktriggered:
		return "IMS voice EPS fallback or RAT fallback triggered"
	case ies.CauseRadioNetworkUpintegrityprotectionnotpossible:
		return "UP integrity protection not possible"
	case ies.CauseRadioNetworkUpconfidentialityprotectionnotpossible:
		return "UP confidentiality protection not possible"
	case ies.CauseRadioNetworkSlicenotsupported:
		return "Slice not supported"
	case ies.CauseRadioNetworkUeinrrcinactivestatenotreachable:
		return "UE in RRC inactive state not reachable"
	case ies.CauseRadioNetworkRedirection:
		return "Redirection"
	case ies.CauseRadioNetworkResourcesnotavailablefortheslice:
		return "Resources not available for the slice"
	case ies.CauseRadioNetworkUemaxintegrityprotecteddataratereason:
		return "UE maximum integrity protected data rate reason"
	case ies.CauseRadioNetworkReleaseduetocndetectedmobility:
		return "Release due to CN detected mobility"
	default:
		return "Unknown cause for radio network"
	}
}

func causeTransportToString(transport *ies.CauseTransport) string {
	switch transport.Value {
	case ies.CauseTransportTransportresourceunavailable:
		return "Transport resource unavailable"
	case ies.CauseTransportUnspecified:
		return "Unspecified cause for transport"
	default:
		return "Unknown cause for transport"
	}
}

func causeNasToString(nas *ies.CauseNas) string {
	switch nas.Value {
	case ies.CauseNasNormalrelease:
		return "Normal release"
	case ies.CauseNasAuthenticationfailure:
		return "Authentication failure"
	case ies.CauseNasDeregister:
		return "Deregister"
	case ies.CauseNasUnspecified:
		return "Unspecified cause for NAS"
	default:
		return "Unknown cause for NAS"
	}
}

func causeProtocolToString(protocol *ies.CauseProtocol) string {
	switch protocol.Value {
	case ies.CauseProtocolTransfersyntaxerror:
		return "Transfer syntax error"
	case ies.CauseProtocolAbstractsyntaxerrorreject:
		return "Abstract syntax error - Reject"
	case ies.CauseProtocolAbstractsyntaxerrorignoreandnotify:
		return "Abstract syntax error - Ignore and notify"
	case ies.CauseProtocolMessagenotcompatiblewithreceiverstate:
		return "Message not compatible with receiver state"
	case ies.CauseProtocolSemanticerror:
		return "Semantic error"
	case ies.CauseProtocolAbstractsyntaxerrorfalselyconstructedmessage:
		return "Abstract syntax error - Falsely constructed message"
	case ies.CauseProtocolUnspecified:
		return "Unspecified cause for protocol"
	default:
		return "Unknown cause for protocol"
	}
}

func causeMiscToString(misc *ies.CauseMisc) string {
	switch misc.Value {
	case ies.CauseMiscControlprocessingoverload:
		return "Control processing overload"
	case ies.CauseMiscNotenoughuserplaneprocessingresources:
		return "Not enough user plane processing resources"
	case ies.CauseMiscHardwarefailure:
		return "Hardware failure"
	case ies.CauseMiscOmintervention:
		return "OM (Operations and Maintenance) intervention"
	case ies.CauseMiscUnknownplmn:
		return "Unknown PLMN (Public Land Mobile Network)"
	case ies.CauseMiscUnspecified:
		return "Unspecified cause for miscellaneous"
	default:
		return "Unknown cause for miscellaneous"
	}
}
