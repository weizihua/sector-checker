package main

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/docker/go-units"
	"github.com/filecoin-project/go-address"
	paramfetch "github.com/filecoin-project/go-paramfetch"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/build"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper/basicfs"
	saproof "github.com/filecoin-project/specs-actors/actors/runtime/proof"
	"github.com/ipfs/go-cid"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

type CheckResults struct {
	SectorSize abi.SectorSize

	// SealingResults []SealingResult

	PostGenerateCandidates  time.Duration
	PostWinningProofCold    time.Duration
	PostWinningProofHot     time.Duration
	VerifyWinningPostCold   time.Duration
	VerifyWinningPostHot    time.Duration
	PostGenerateCandidatesM time.Duration
	VerifyWinningPostColdM  time.Duration

	PostWindowProofCold  time.Duration
	PostWindowProofHot   time.Duration
	VerifyWindowPostCold time.Duration
	VerifyWindowPostHot  time.Duration
}

var sealBenchCmd = &cli.Command{
	Name: "checking",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "storage-dir",
			Value: "~/.lotus-checking",
			Usage: "Path to the storage directory that will store sectors long term",
		},
		&cli.StringFlag{
			Name:  "sector-size",
			Value: "32GiB",
			Usage: "size of the sectors in bytes, i.e. 512MiB, 32GiB",
		},
		&cli.StringFlag{
			Name:  "sectors-file",
			Value: "sectors.txt",
			Usage: "absolute path file. contains line number, line cidcommr, line number...",
		},
		&cli.StringFlag{
			Name:  "sectors-file-only-number",
			Value: "",
			Usage: "absolute path file. contains line number, line cidcommr, line number...",
		},
		&cli.StringFlag{
			Name:  "sectors-file-with-cidcommr",
			Value: "sectors.txt",
			Usage: "aabsolute path file. contains line number, line cidcommr, line number..., each in a line",
		},
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the checking",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t010010",
		},
		&cli.IntFlag{
			Name:  "number",
			Value: 1,
		},
		&cli.StringFlag{
			Name:  "cidcommr",
			Usage: "CIDcommR,  eg/default.  bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z",
			Value: "bagboea4b5abcbkyyzhl37s5kyjjegeysedpczhija7cczazapavjejbppck57b2z",
		},
	},
	Action: func(c *cli.Context) error {
		// policy.AddSupportedProofTypes(abi.RegisteredSealProof_StackedDrg2KiBV1)

		if c.Bool("no-gpu") {
			err := os.Setenv("BELLMAN_NO_GPU", "1")
			if err != nil {
				return xerrors.Errorf("setting no-gpu flag: %w", err)
			}
		}

		var sbdir string

		sdir, err := homedir.Expand(c.String("storage-dir"))
		if err != nil {
			log.Errorf("===homedir.Expand==: %+v", err)
			return err
		}
		log.Debugf("===storage-dir==: %+v", sdir)

		err = os.MkdirAll(sdir, 0775) //nolint:gosec
		if err != nil {
			return xerrors.Errorf("creating sectorbuilder dir: %w", err)
		}

		// defer func() {
		// }()

		// st, err := os.Stat(sdir)
		// if err != nil {
		// 	return xerrors.Errorf("File (%s) does not exist: %w", sdir, err)
		// }
		// log.Debugf("===os.Stat==: %+v", st)

		// f, err := os.Open(sdir)
		// if err != nil {
		// 	return xerrors.Errorf("opening backup file: %w", err)
		// }
		// defer f.Close() // nolint:errcheck
		// log.Debugf("===os.Open==: %+v", f)
		// log.Info("Checking if repo exists")

		sbdir = sdir

		// miner address
		maddr, err := address.NewFromString(c.String("miner-addr"))
		if err != nil {
			return err
		}
		log.Infof("miner maddr: ", maddr)
		amid, err := address.IDFromAddress(maddr)
		if err != nil {
			return err
		}
		log.Infof("miner amid: ", amid)
		mid := abi.ActorID(amid)
		log.Infof("miner mid: ", mid)

		// sector size
		sectorSizeInt, err := units.RAMInBytes(c.String("sector-size"))
		if err != nil {
			return err
		}
		sectorSize := abi.SectorSize(sectorSizeInt)

		spt, err := ffiwrapper.SealProofTypeFromSectorSize(sectorSize)
		if err != nil {
			return err
		}

		cfg := &ffiwrapper.Config{
			SealProofType: spt,
		}

		if err := paramfetch.GetParams(lcli.ReqContext(c), build.ParametersJSON(), uint64(sectorSize)); err != nil {
			return xerrors.Errorf("getting params: %w", err)
		}

		sbfs := &basicfs.Provider{
			Root: sbdir,
		}

		sb, err := ffiwrapper.New(sbfs, cfg)
		if err != nil {
			return err
		}

		sealedSectors := getSectorsInfo(c.String("sectors-file"), sb.SealProofType())

		var challenge [32]byte
		rand.Read(challenge[:])

		windowpostStart := time.Now()

		log.Info("computing window post snark (cold)")
		wproof1, ps, err := sb.GenerateWindowPoSt(context.TODO(), mid, sealedSectors, challenge[:])
		if err != nil {
			return err
		}
		windowpost1 := time.Now()

		wpvi1 := saproof.WindowPoStVerifyInfo{
			Randomness:        challenge[:],
			Proofs:            wproof1,
			ChallengedSectors: sealedSectors,
			Prover:            mid,
		}

		log.Info("generate window PoSt skipped sectors", "sectors", ps, "error", err)

		ok, err := ffiwrapper.ProofVerifier.VerifyWindowPoSt(context.TODO(), wpvi1)
		if err != nil {
			return err
		}
		if !ok {
			log.Error("post verification failed")
		}
		verifyWindowpost1 := time.Now()

		// bo := CheckResults{
		// 	SectorSize: sectorSize,
		// 	// SealingResults: sealTimings,
		// }

		PostGenerateCandidates := windowpost1.Sub(windowpostStart)
		VerifyWinningPostCold := verifyWindowpost1.Sub(windowpost1)

		// bo.PostGenerateCandidatesM = bo.PostGenerateCandidates.Truncate(time.Millisecond * 100)
		// bo.VerifyWinningPostColdM = bo.VerifyWinningPostCold.Truncate(time.Millisecond * 100)

		fmt.Printf("PostGenerateCandidates == %+v:\n", PostGenerateCandidates)
		fmt.Printf("VerifyWinningPostCold == %+v:\n", VerifyWinningPostCold)

		// data, err := json.MarshalIndent(bo, "", "  ")
		// if err != nil {
		// 	return err
		// }

		// fmt.Println(string(data))

		return nil
	},
}

func getSectorsInfo(filePath string, proofType abi.RegisteredSealProof) []saproof.SectorInfo {

	sealedSectors := make([]saproof.SectorInfo, 0)

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return sealedSectors
	}
	log.Debugf("===filePath==: %+v", filePath)
	log.Debugf("===os.Open()==: %+v", file)
	scanner := bufio.NewScanner(file)
	log.Debugf("===bufio.NewScanner()==: %+v", scanner)
	for scanner.Scan() {
		sectorIndex := scanner.Text()
		log.Debugf("===oscanner.Text()=254=: %+v", sectorIndex)
		index, error := strconv.Atoi(sectorIndex)
		log.Debugf("===strconv.Atoi()=256=: %+v", index)
		if error != nil {
			fmt.Println("error")
			break
		}

		scanner.Scan()
		cidStr := scanner.Text()
		log.Debugf("===oscanner.Text()==: %+v", cidStr)
		ccid, err := cid.Decode(cidStr)
		log.Debugf("===cid.Decode()==: %+v", ccid)
		if err != nil {
			log.Infof("cid error, ignore sectors after this: %d, %s", uint64(index), err)
			return sealedSectors
		}

		var sector saproof.SectorInfo
		sector.SealProof = proofType
		sector.SectorNumber = abi.SectorNumber(uint64(index))
		sector.SealedCID = ccid

		sealedSectors = append(sealedSectors, sector)

		log.Infof("id: ", sector.SectorNumber)
		log.Infof("cid: ", sector.SealedCID)

	}

	fmt.Println("sector length", len(sealedSectors))
	return sealedSectors
}
