package utils

import (
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	"io"
	"os"

	"github.com/ava-labs/ava-sim/constants"

	"github.com/ava-labs/avalanchego/ids"
	avalancheContants "github.com/ava-labs/avalanchego/utils/constants"
	"github.com/ava-labs/avalanchego/utils/hashing"
)

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// Grant permission to copy
	if err := os.Chmod(dst, constants.FilePerms); err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func LoadNodeID(stakeCert []byte) (string, error) {
	block, _ := pem.Decode(stakeCert)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("%w: problem parsing staking certificate", err)
	}

	id, err := ids.ToShortID(hashing.PubkeyBytesToAddress(cert.Raw))
	if err != nil {
		return "", fmt.Errorf("%w: problem deriving staker ID from certificate", err)
	}

	return id.PrefixedString(avalancheContants.NodeIDPrefix), nil
}
