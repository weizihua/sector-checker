package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"time"

	"github.com/filecoin-project/go-address"
	paramfetch "github.com/filecoin-project/go-paramfetch"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/types"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/extern/sector-storage/ffiwrapper"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

type ParCfg struct {
	PreCommit1 int
	PreCommit2 int
	Commit     int
}

var proveCmd = &cli.Command{
	Name:  "prove",
	Usage: "Benchmark a proof computation",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "no-gpu",
			Usage: "disable gpu usage for the benchmark run",
		},
		&cli.StringFlag{
			Name:  "miner-addr",
			Usage: "pass miner address (only necessary if using existing sectorbuilder)",
			Value: "t01000",
		},
	},
	Action: func(c *cli.Context) error {
		if c.Bool("no-gpu") {
			err := os.Setenv("BELLMAN_NO_GPU", "1")
			if err != nil {
				return xerrors.Errorf("setting no-gpu flag: %w", err)
			}
		}

		if !c.Args().Present() {
			return xerrors.Errorf("Usage: lotus-bench prove [input.json]")
		}

		inb, err := ioutil.ReadFile(c.Args().First())
		if err != nil {
			return xerrors.Errorf("reading input file: %w", err)
		}

		var c2in Commit2In
		if err := json.Unmarshal(inb, &c2in); err != nil {
			return xerrors.Errorf("unmarshalling input file: %w", err)
		}

		if err := paramfetch.GetParams(lcli.ReqContext(c), build.ParametersJSON(), c2in.SectorSize); err != nil {
			return xerrors.Errorf("getting params: %w", err)
		}

		maddr, err := address.NewFromString(c.String("miner-addr"))
		if err != nil {
			return err
		}
		mid, err := address.IDFromAddress(maddr)
		if err != nil {
			return err
		}

		spt, err := ffiwrapper.SealProofTypeFromSectorSize(abi.SectorSize(c2in.SectorSize))
		if err != nil {
			return err
		}

		cfg := &ffiwrapper.Config{
			SealProofType: spt,
		}

		sb, err := ffiwrapper.New(nil, cfg)
		if err != nil {
			return err
		}

		start := time.Now()

		proof, err := sb.SealCommit2(context.TODO(), abi.SectorID{Miner: abi.ActorID(mid), Number: abi.SectorNumber(c2in.SectorNum)}, c2in.Phase1Out)
		if err != nil {
			return err
		}

		sealCommit2 := time.Now()

		fmt.Printf("proof: %x\n", proof)

		fmt.Printf("----\nresults (v27) (%d)\n", c2in.SectorSize)
		dur := sealCommit2.Sub(start)

		fmt.Printf("seal: commit phase 2: %s (%s)\n", dur, bps(abi.SectorSize(c2in.SectorSize), dur))
		return nil
	},
}

func bps(data abi.SectorSize, d time.Duration) string {
	bdata := new(big.Int).SetUint64(uint64(data))
	bdata = bdata.Mul(bdata, big.NewInt(time.Second.Nanoseconds()))
	bps := bdata.Div(bdata, big.NewInt(d.Nanoseconds()))
	return types.SizeStr(types.BigInt{Int: bps}) + "/s"
}
