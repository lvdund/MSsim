package context

import "github.com/reogac/nas"

func cause5GMMToString(cause5GMM uint8) string {
	switch cause5GMM {
	case nas.Cause5GMMIllegalUE:
		return "Illegal UE"
	case nas.Cause5GMMPEINotAccepted:
		return "PEI not accepted"
	case nas.Cause5GMMIllegalME:
		return "5GS services not allowed"
	case nas.Cause5GMM5GSServicesNotAllowed:
		return "5GS services not allowed"
	case nas.Cause5GMMUEIdentityCannotBeDerivedByTheNetwork:
		return "UE identity cannot be derived by the network"
	case nas.Cause5GMMImplicitlyDeregistered:
		return "Implicitly de-registered"
	case nas.Cause5GMMPLMNNotAllowed:
		return "PLMN not allowed"
	case nas.Cause5GMMTrackingAreaNotAllowed:
		return "Tracking area not allowed"
	case nas.Cause5GMMRoamingNotAllowedInThisTrackingArea:
		return "Roaming not allowed in this tracking area"
	case nas.Cause5GMMNoSuitableCellsInTrackingArea:
		return "No suitable cells in tracking area"
	case nas.Cause5GMMMACFailure:
		return "MAC failure"
	case nas.Cause5GMMSynchFailure:
		return "Synch failure"
	case nas.Cause5GMMCongestion:
		return "Congestion"
	case nas.Cause5GMMUESecurityCapabilitiesMismatch:
		return "UE security capabilities mismatch"
	case nas.Cause5GMMSecurityModeRejectedUnspecified:
		return "Security mode rejected, unspecified"
	case nas.Cause5GMMNon5GAuthenticationUnacceptable:
		return "Non-5G authentication unacceptable"
	case nas.Cause5GMMN1ModeNotAllowed:
		return "N1 mode not allowed"
	case nas.Cause5GMMRestrictedServiceArea:
		return "Restricted service area"
	case nas.Cause5GMMLADNNotAvailable:
		return "LADN not available"
	case nas.Cause5GMMMaximumNumberOfPDUSessionsReached:
		return "Maximum number of PDU sessions reached"
	case nas.Cause5GMMInsufficientResourcesForSpecificSliceAndDNN:
		return "Insufficient resources for specific slice and DNN"
	case nas.Cause5GMMInsufficientResourcesForSpecificSlice:
		return "Insufficient resources for specific slice"
	case nas.Cause5GMMngKSIAlreadyInUse:
		return "ngKSI already in use"
	case nas.Cause5GMMNon3GPPAccessTo5GCNNotAllowed:
		return "Non-3GPP access to 5GCN not allowed"
	case nas.Cause5GMMServingNetworkNotAuthorized:
		return "Serving network not authorized"
	case nas.Cause5GMMPayloadWasNotForwarded:
		return "Payload was not forwarded"
	case nas.Cause5GMMDNNNotSupportedOrNotSubscribedInTheSlice:
		return "DNN not supported or not subscribed in the slice"
	case nas.Cause5GMMInsufficientUserPlaneResourcesForThePDUSession:
		return "Insufficient user-plane resources for the PDU session"
	case nas.Cause5GMMSemanticallyIncorrectMessage:
		return "Semantically incorrect message"
	case nas.Cause5GMMInvalidMandatoryInformation:
		return "Invalid mandatory information"
	case nas.Cause5GMMMessageTypeNonExistentOrNotImplemented:
		return "Message type non-existent or not implementedE"
	case nas.Cause5GMMMessageTypeNotCompatibleWithTheProtocolState:
		return "Message type not compatible with the protocol state"
	case nas.Cause5GMMInformationElementNonExistentOrNotImplemented:
		return "Information element non-existent or not implemented"
	case nas.Cause5GMMConditionalIEError:
		return "Conditional IE error"
	case nas.Cause5GMMMessageNotCompatibleWithTheProtocolState:
		return "Message not compatible with the protocol state"
	case nas.Cause5GMMProtocolErrorUnspecified:
		return "Protocol error, unspecified. Please share the pcap with mssim@hpe.com."
	default:
		return "Protocol error, unspecified. Please share the pcap with mssim@hpe.com."
	}
}

func cause5GSMToString(causeValue uint8) string {
	switch causeValue {
	case nas.Cause5GSMInsufficientResources:
		return "Insufficient Ressources"
	case nas.Cause5GSMMissingOrUnknownDNN:
		return "Missing or Unknown DNN"
	case nas.Cause5GSMUnknownPDUSessionType:
		return "Unknown PDU Session Type"
	case nas.Cause5GSMUserAuthenticationOrAuthorizationFailed:
		return "User authentification or authorization failed"
	case nas.Cause5GSMRequestRejectedUnspecified:
		return "Request rejected, unspecified"
	case nas.Cause5GSMServiceOptionTemporarilyOutOfOrder:
		return "Service option temporarily out of order."
	case nas.Cause5GSMPTIAlreadyInUse:
		return "PTI already in use"
	case nas.Cause5GSMRegularDeactivation:
		return "Regular deactivation"
	case nas.Cause5GSMReactivationRequested:
		return "Reactivation requested"
	case nas.Cause5GSMInvalidPDUSessionIdentity:
		return "Invalid PDU session identity"
	case nas.Cause5GSMSemanticErrorsInPacketFilter:
		return "Semantic errors in packet filter(s)"
	case nas.Cause5GSMSyntacticalErrorInPacketFilter:
		return "Syntactical error in packet filter(s)"
	case nas.Cause5GSMOutOfLADNServiceArea:
		return "Out of LADN service area"
	case nas.Cause5GSMPTIMismatch:
		return "PTI mismatch"
	case nas.Cause5GSMPDUSessionTypeIPv4OnlyAllowed:
		return "PDU session type IPv4 only allowed"
	case nas.Cause5GSMPDUSessionTypeIPv6OnlyAllowed:
		return "PDU session type IPv6 only allowed"
	case nas.Cause5GSMPDUSessionDoesNotExist:
		return "PDU session does not exist"
	case nas.Cause5GSMInsufficientResourcesForSpecificSliceAndDNN:
		return "Insufficient resources for specific slice and DNN"
	case nas.Cause5GSMNotSupportedSSCMode:
		return "Not supported SSC mode"
	case nas.Cause5GSMInsufficientResourcesForSpecificSlice:
		return "Insufficient resources for specific slice"
	case nas.Cause5GSMMissingOrUnknownDNNInASlice:
		return "Missing or unknown DNN in a slice"
	case nas.Cause5GSMInvalidPTIValue:
		return "Invalid PTI value"
	case nas.Cause5GSMMaximumDataRatePerUEForUserPlaneIntegrityProtectionIsTooLow:
		return "Maximum data rate per UE for user-plane integrity protection is too low"
	case nas.Cause5GSMSemanticErrorInTheQoSOperation:
		return "Semantic error in the QoS operation"
	case nas.Cause5GSMSyntacticalErrorInTheQoSOperation:
		return "Syntactical error in the QoS operation"
	case nas.Cause5GSMInvalidMappedEPSBearerIdentity:
		return "Invalid mapped EPS bearer identity"
	case nas.Cause5GSMSemanticallyIncorrectMessage:
		return "Semantically incorrect message"
	case nas.Cause5GSMInvalidMandatoryInformation:
		return "Invalid mandatory information"
	case nas.Cause5GSMMessageTypeNonExistentOrNotImplemented:
		return "Message type non-existent or not implemented"
	case nas.Cause5GSMMessageTypeNotCompatibleWithTheProtocolState:
		return "Message type not compatible with the protocol state"
	case nas.Cause5GSMInformationElementNonExistentOrNotImplemented:
		return "Information element non-existent or not implemented"
	case nas.Cause5GSMConditionalIEError:
		return "Conditional IE error"
	case nas.Cause5GSMMessageNotCompatibleWithTheProtocolState:
		return "Message not compatible with the protocol state"
	case nas.Cause5GSMProtocolErrorUnspecified:
		return "Protocol error, unspecified. Please open an issue on Github with pcap."
	default:
		return "Service option temporarily out of order."
	}
}
