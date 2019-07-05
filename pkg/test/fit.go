package test

import (
	"fmt"
	"os"

	"github.com/9elements/txt-suite/pkg/api"
)

// 16MiB
const FITSize int64 = 16 * 1024 * 1024
const FourGiB int64 = 0x100000000
const ResetVector = 0xFFFFFFF0

var (
	fitImage []byte
	TestsFIT = [...]Test{
		Test{
			name:     "Has FIT",
			required: true,
			function: Test22HasFIT,
		},
		Test{
			name:     "FIT has an BIOS ACM entry",
			required: true,
			function: Test23HasBIOSACM,
		},
		Test{
			name:     "FIT has a initial bootblock entry",
			required: true,
			function: Test24HasIBB,
		},
		Test{
			name:     "FIT has a LCP Policy entry",
			required: true,
			function: Test25HasBIOSPolicy,
		},
		Test{
			name:     "Initial bootblock covers reset vector",
			required: true,
			function: Test26IBBCoversResetVector,
		},
		Test{
			name:     "Initial bootblock does not overlap",
			required: true,
			function: Test27NoIBBOverlap,
		},
		Test{
			name:     "BIOS ACM does not overlap",
			required: true,
			function: Test28NoBIOSACMOverlap,
		},
		Test{
			name:     "Initial bootblock and BIOS ACM is below 4GiB",
			required: true,
			function: Test29BIOSACMIsBelow4G,
		},
		Test{
			name:     "LCP Policy does not disable Intel TXT",
			required: true,
			function: Test30PolicyAllowsTXT,
		},
	}
)

func LoadFITFromMemory() error {
	var buf api.Uint8

	for i := int64(0); i < FITSize; i += 1 {
		err := api.ReadPhys(FourGiB-FITSize+i, &buf)
		if err != nil {
			return err
		}

		fitImage = append(fitImage, byte(buf))
	}

	return nil
}

func LoadFITFromFile(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}

	defer fd.Close()

	fitImage = make([]byte, FITSize)
	_, err = fd.ReadAt(fitImage, FourGiB-FITSize)
	if err != nil {
		return err
	}

	return nil
}

func Test22HasFIT() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	// XXX: verify checksum

	return len(fit) > 0, nil
}

func Test23HasBIOSACM() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	count := 0
	for _, ent := range fit {
		if ent.Type() == api.StartUpACMod {
			count += 1
		}
	}

	return count == 1, nil
}

func Test24HasIBB() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for _, ent := range fit {
		if ent.Type() == api.BIOSStartUpMod {
			return true, nil
		}
	}

	return false, nil
}

func Test25HasBIOSPolicy() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for _, ent := range fit {
		if ent.Type() == api.BIOSPolicyRec {
			return true, nil
		}
	}

	return false, nil
}

func Test26IBBCoversResetVector() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for _, ent := range fit {
		if ent.Type() == api.BIOSStartUpMod {
			coversRv := ent.Address <= ResetVector && ent.Address+uint64(ent.Size()) >= ResetVector+4

			if coversRv {
				return true, nil
			}
		}
	}

	return false, nil
}

func Test27NoIBBOverlap() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for i, ent1 := range fit {
		if ent1.Type() == api.BIOSStartUpMod {
			for j, ent2 := range fit {
				if i < j && ent2.Type() == api.BIOSStartUpMod {
					a := ent1.Address > ent2.Address+uint64(ent2.Size())
					b := ent2.Address > ent1.Address+uint64(ent1.Size())

					if !a && !b {
						return false, nil
					}
				}
			}
		}
	}

	return true, nil
}

func Test28NoBIOSACMOverlap() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for i, ent1 := range fit {
		if ent1.Type() == api.BIOSStartUpMod {
			for j, ent2 := range fit {
				if i < j && ent2.Type() == api.StartUpACMod {
					a := ent1.Address > ent2.Address+uint64(ent2.Size())
					b := ent2.Address > ent1.Address+uint64(ent1.Size())

					if !a && !b {
						return false, nil
					}
				}
			}
		}
	}

	return true, nil
}

func Test29BIOSACMIsBelow4G() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for _, ent := range fit {
		if ent.Type() == api.StartUpACMod {
			if ent.Address+uint64(ent.Size()) > uint64(FourGiB) {
				return false, nil
			}
		}
	}

	return true, nil
}

func Test30PolicyAllowsTXT() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for _, ent := range fit {
		if ent.Type() == api.TXTPolicyRec {
			switch ent.Version {
			case 0:
				return false, fmt.Errorf("Indexed IO type pointer are not supported")
			case 1:
				var b api.Uint8

				err = api.ReadPhys(int64(ent.Address), &b)
				if err != nil {
					return false, err
				}

				return b&1 != 0, nil
			default:
				return false, fmt.Errorf("Unknown TXT policy record version %d", ent.Version)
			}
		}
	}

	// No record means TXT is enabled
	return true, nil
}

func Test31BIOSACMValid() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	acm, _, _, _, err := biosACM(fit)

	return acm != nil, err
}

func Test32BIOSACMSizeCorrect() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	acm, _, _, _, err := biosACM(fit)
	if err != nil {
		return false, err
	}

	return acm.HeaderLen%64 == 0, nil
}

func Test33BIOSACMAlignmentCorrect() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	for _, ent := range fit {
		if ent.Type() == api.StartUpACMod {
			return ent.Address%(128*1024) == 0, nil
		}
	}

	return false, fmt.Errorf("no BIOS ACM in FIT")
}

func Test34BIOSACMMatchesChipset() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	acm, chp, _, _, err := biosACM(fit)
	if err != nil {
		return false, err
	}

	txt, err := api.ReadTXTRegs()
	if err != nil {
		return false, err
	}

	for _, ch := range chp.IDList {
		a := ch.VendorID == txt.Vid
		b := ch.DeviceID == txt.Did

		if a && b {
			if acm.Flags&1 != 0 {
				if ch.RevisionID&txt.Rid == txt.Rid {
					return true, nil
				}
			} else {
				if ch.RevisionID == txt.Rid {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func Test35BIOSACMMatchesCPU() (bool, error) {
	if len(fitImage) == 0 {
		return false, fmt.Errorf("No FIT image loaded")
	}

	fit, err := api.ExtractFit(fitImage)
	if err != nil {
		return false, err
	}

	_, _, cpus, _, err := biosACM(fit)
	if err != nil {
		return false, err
	}

	// IA32_PLATFORM_ID
	platform, err := api.IA32PlatformID()
	if err != nil {
		return false, err
	}

	fms := api.FamilyModelStepping()

	for _, cpu := range cpus.IDList {
		a := fms&cpu.FMSMask == cpu.FMS
		b := platform&cpu.PlatformMask == cpu.PlatformID

		if a && b {
			return true, nil
		}
	}

	return false, nil
}

func biosACM(fit []api.FitEntry) (*api.ACM, *api.Chipsets, *api.Processors, *api.TPMs, error) {
	for _, ent := range fit {
		if ent.Type() == api.StartUpACMod {
			buf := make([]byte, 224*4)

			err := api.ReadPhysBuf(int64(ent.Address), buf)
			if err != nil {
				return nil, nil, nil, nil, err
			}

			return api.ParseACM(buf)
		}
	}

	return nil, nil, nil, nil, fmt.Errorf("no BIOS ACM in FIT")
}
