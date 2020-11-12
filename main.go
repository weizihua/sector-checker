package main

import (
	"os"

	//	"path/filepath"

	logging "github.com/ipfs/go-log/v2"

	//	"github.com/minio/blake2b-simd"

	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/go-state-types/abi"

	//	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	//	"github.com/filecoin-project/specs-storage/storage"

	//	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
)

var log = logging.Logger("sector-checker")

type Commit2In struct {
	SectorNum  int64
	Phase1Out  []byte
	SectorSize uint64
}

func main() {
	logging.SetLogLevel("*", "DEBUG")

	log.Info("Starting sector-checker")

	miner.SupportedProofTypes[abi.RegisteredSealProof_StackedDrg2KiBV1] = struct{}{}

	app := &cli.App{
		Name:    "sector-sanity-check",
		Usage:   "check window post",
		Version: build.UserVersion(),
		Commands: []*cli.Command{
			proveCmd,
			sealBenchCmd,
			importBenchCmd,
			sectorsCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}
