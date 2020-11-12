package main

import (
	"fmt"
	"os"

	//	"path/filepath"
	logging "github.com/ipfs/go-log/v2"
	//	"github.com/minio/blake2b-simd"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/urfave/cli/v2"

	//	"github.com/filecoin-project/lotus/extern/sector-storage/stores"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	//	"github.com/filecoin-project/specs-storage/storage"
	//	lapi "github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/node/repo"
)

var log = logging.Logger("sector-checker")

// const FlagMinerRepo = "miner-repo"
const FlagMinerRepo = "storage-dir"

// TODO remove after deprecation period
const FlagMinerRepoDeprecation = "storagerepo"

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
		Name:                 "sector-sanity-check",
		Usage:                "check window post",
		Version:              build.UserVersion(),
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    FlagMinerRepo,
				Aliases: []string{FlagMinerRepoDeprecation},
				EnvVars: []string{"LOTUS_MINER_PATH", "LOTUS_STORAGE_PATH"},
				Value:   "~/.lotus-checking", // TODO: Consider XDG_DATA_HOME
				Usage:   fmt.Sprintf("Specify miner repo path. flag(%s) and env(LOTUS_STORAGE_PATH) are DEPRECATION, will REMOVE SOON", FlagMinerRepoDeprecation),
			},

			// &cli.StringFlag{
			// 	Name:  "storage-dir",
			// 	Value: "~/.lotus-checking",
			// 	Usage: "Path to the storage directory that will store sectors long term",
			// },
		},
		Commands: []*cli.Command{
			proveCmd,
			sealBenchCmd,
			importBenchCmd,
			sectorsCmd,
		},
	}

	app.Setup()

	// repo.FullNode: "FULLNODE_API_INFO"
	// repo.StorageMiner: "MINER_API_INFO"
	// repo.Worker: "WORKER_API_INFO"
	app.Metadata["repoType"] = repo.StorageMiner

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%+v", err)
		return
	}
}
