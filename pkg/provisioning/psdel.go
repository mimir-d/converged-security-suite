package provisioning

import (
	"fmt"
	"io"

	"github.com/google/go-tpm/tpmutil"

	tpm2 "github.com/google/go-tpm/tpm2"
)

// DeletePSindexTPM20 deletes the PS index on TPM 2.0
func DeletePSindexTPM20(rw io.ReadWriter, delHash, writeHash []byte) error {
	zeroHash := make([]byte, 32)
	delPol, err := constructDelBranch(rw, delHash, zeroHash)
	if err != nil {
		return err
	}
	writePol, err := constructWriteBranch(rw, writeHash, zeroHash)
	if err != nil {
		return err
	}
	// Policy Session for authorizing NV access to PS index
	sessIndex, _, err := tpm2.StartAuthSession(rw, tpm2.HandleNull, tpm2.HandleNull, make([]byte, 16), nil, tpm2.SessionPolicy, tpm2.AlgNull, tpm2.AlgSHA256)
	if err != nil {
		return err
	}
	defer tpm2.FlushContext(rw, sessIndex)

	or1 := tpm2.TPMLDigest{Digests: []tpmutil.U16Bytes{delHash, zeroHash}}
	or2 := tpm2.TPMLDigest{Digests: []tpmutil.U16Bytes{delPol, writePol}}

	err = tpm2.PolicyOr(rw, sessIndex, or1)
	if err != nil {
		return fmt.Errorf("PolicyOr1 failed with: %v", err)
	}

	err = tpm2.PolicyCommandCode(rw, sessIndex, tpm2.CmdNVUndefineSpaceSpecial)
	if err != nil {
		return fmt.Errorf("PolicyCommandCode failed with: %v", err)
	}

	err = tpm2.PolicyOr(rw, sessIndex, or2)
	if err != nil {
		return fmt.Errorf("PolicyOr2 failed with: %v", err)
	}

	err = tpm2.NVUndefineSpaceSpecial(rw, tpm2PSIndexDef.NVIndex, sessIndex, tpm2.EmptyAuth)
	if err != nil {
		return fmt.Errorf("NVUndefineSpaceSpecial() failed: %v", err)
	}

	return err
}

// DeletePSIndexTPM12 deletes the PS index on TPM 1.2
func DeletePSIndexTPM12(rw io.ReadWriter) error {
	return fmt.Errorf("Not implemented yet")
}
