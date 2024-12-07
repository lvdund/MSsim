package auth

import (
	//	"github.com/free5gc/util/ueauth"
	"github.com/reogac/nas"
)

// Algorithm key Derivation function defined in TS 33.501 Annex A.9
func AlgorithmKeyDerivation(cipheringAlg uint8, kamf []byte, knasEnc *[16]uint8, integrityAlg uint8, knasInt *[16]uint8) error {
	/*
		// Security Key
		P0 := []byte{security.NNASEncAlg}
		L0 := ueauth.KDFLen(P0)
		P1 := []byte{cipheringAlg}
		L1 := ueauth.KDFLen(P1)

		kenc, err := ueauth.GetKDFValue(kamf, ueauth.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
		if err != nil {
			return err
		}
		copy(knasEnc[:], kenc[16:32])

		// Integrity Key
		P0 = []byte{security.NNASIntAlg}
		L0 = ueauth.KDFLen(P0)
		P1 = []byte{integrityAlg}
		L1 = ueauth.KDFLen(P1)

		kint, err := ueauth.GetKDFValue(kamf, ueauth.FC_FOR_ALGORITHM_KEY_DERIVATION, P0, L0, P1, L1)
		if err != nil {
			return err
		}
		copy(knasInt[:], kint[16:32])
	*/
	return nil
}

func SelectAlgorithms(secCap *nas.UeSecurityCapability) (intergritygAlgorithm uint8, cipheringAlgorithm uint8) {
	/*
		//TODO
		// set the algorithms of integrity
		if secCap.GetIA0_5G() == 1 {
			intergritygAlgorithm = security.AlgIntegrity128NIA0
		} else if secCap.GetIA1_128_5G() == 1 {
			intergritygAlgorithm = security.AlgIntegrity128NIA1
		} else if secCap.GetIA2_128_5G() == 1 {
			intergritygAlgorithm = security.AlgIntegrity128NIA2
		} else if secCap.GetIA3_128_5G() == 1 {
			intergritygAlgorithm = security.AlgIntegrity128NIA3
		}

		// set the algorithms of ciphering
		if secCap.GetEA0_5G() == 1 {
			cipheringAlgorithm = security.AlgCiphering128NEA0
		} else if secCap.GetEA1_128_5G() == 1 {
			cipheringAlgorithm = security.AlgCiphering128NEA1
		} else if secCap.GetEA2_128_5G() == 1 {
			cipheringAlgorithm = security.AlgCiphering128NEA2
		} else if secCap.GetEA3_128_5G() == 1 {
			cipheringAlgorithm = security.AlgCiphering128NEA3
		}
	*/

	return
}
