package main

import (
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/filecoin-project/go-state-types/abi"
	lcli "github.com/filecoin-project/lotus/cli"
	sealing "github.com/filecoin-project/lotus/extern/storage-sealing"
	"github.com/filecoin-project/lotus/lib/tablewriter"
	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"
)

type stateMeta struct {
	i     int
	col   color.Attribute
	state sealing.SectorState
}

var stateOrder = map[sealing.SectorState]stateMeta{}
var stateList = []stateMeta{
	{col: 39, state: "Total"},
	{col: color.FgGreen, state: sealing.Proving},

	{col: color.FgBlue, state: sealing.Empty},
	{col: color.FgBlue, state: sealing.WaitDeals},

	{col: color.FgRed, state: sealing.UndefinedSectorState},
	{col: color.FgYellow, state: sealing.Packing},
	{col: color.FgYellow, state: sealing.PreCommit1},
	{col: color.FgYellow, state: sealing.PreCommit2},
	{col: color.FgYellow, state: sealing.PreCommitting},
	{col: color.FgYellow, state: sealing.PreCommitWait},
	{col: color.FgYellow, state: sealing.WaitSeed},
	{col: color.FgYellow, state: sealing.Committing},
	{col: color.FgYellow, state: sealing.SubmitCommit},
	{col: color.FgYellow, state: sealing.CommitWait},
	{col: color.FgYellow, state: sealing.FinalizeSector},

	{col: color.FgCyan, state: sealing.Removing},
	{col: color.FgCyan, state: sealing.Removed},

	{col: color.FgRed, state: sealing.FailedUnrecoverable},
	{col: color.FgRed, state: sealing.SealPreCommit1Failed},
	{col: color.FgRed, state: sealing.SealPreCommit2Failed},
	{col: color.FgRed, state: sealing.PreCommitFailed},
	{col: color.FgRed, state: sealing.ComputeProofFailed},
	{col: color.FgRed, state: sealing.CommitFailed},
	{col: color.FgRed, state: sealing.PackingFailed},
	{col: color.FgRed, state: sealing.FinalizeFailed},
	{col: color.FgRed, state: sealing.Faulty},
	{col: color.FgRed, state: sealing.FaultReported},
	{col: color.FgRed, state: sealing.FaultedFinal},
	{col: color.FgRed, state: sealing.RemoveFailed},
	{col: color.FgRed, state: sealing.DealsExpired},
	{col: color.FgRed, state: sealing.RecoverDealIDs},
}

var sectorsCmd = &cli.Command{
	Name:  "sectors",
	Usage: "interact with sector store",
	Subcommands: []*cli.Command{
		// sectorsStatusCmd,
		sectorsListCmd,
		// sectorsRefsCmd,
		// sectorsUpdateCmd,
		// sectorsPledgeCmd,
		// sectorsRemoveCmd,
		// sectorsMarkForUpgradeCmd,
		// sectorsStartSealCmd,
		// sectorsSealDelayCmd,
		// sectorsCapacityCollateralCmd,
	},
}

var sectorsListCmd = &cli.Command{
	Name:  "list",
	Usage: "List sectors",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "show-removed",
			Usage: "show removed sectors",
		},
		&cli.BoolFlag{
			Name:    "color",
			Aliases: []string{"c"},
			Value:   true,
		},
		&cli.BoolFlag{
			Name:  "fast",
			Usage: "don't show on-chain info for better performance",
		},
	},
	Action: func(cctx *cli.Context) error {
		color.NoColor = !cctx.Bool("color")
		logging.SetLogLevel("rpc", "INFO")

		nodeApi, closer, err := lcli.GetStorageMinerAPI(cctx)
		if err != nil {
			return err
		}
		defer closer()

		fullApi, closer2, err := lcli.GetFullNodeAPI(cctx) // TODO: consider storing full node address in config
		if err != nil {
			return err
		}
		defer closer2()

		ctx := lcli.ReqContext(cctx)

		list, err := nodeApi.SectorsList(ctx)
		if err != nil {
			return err
		}
		log.Debugf("SectorsList: ", len(list))

		maddr, err := nodeApi.ActorAddress(ctx)
		if err != nil {
			return err
		}
		log.Debugf("ActorAddress: ", maddr)

		head, err := fullApi.ChainHead(ctx)
		if err != nil {
			return err
		}
		log.Debugf("ChainHead: ", head)

		activeSet, err := fullApi.StateMinerActiveSectors(ctx, maddr, head.Key())
		if err != nil {
			return err
		}
		log.Debugf("activeSet: ", len(activeSet))

		// secCounts, err := fullApi.StateMinerSectorCount(ctx, maddr, types.EmptyTSK)
		// if err != nil {
		// 	return err
		// }
		// log.Debugf("activeSet: ", len(secCounts))
		// proving := secCounts.Active + secCounts.Faulty
		// nfaults := secCounts.Faulty

		activeIDs := make(map[abi.SectorNumber]struct{}, len(activeSet))
		log.Debugf("activeIDs: ", activeIDs)

		for _, info := range activeSet {
			activeIDs[info.SectorNumber] = struct{}{}
		}

		sset, err := fullApi.StateMinerSectors(ctx, maddr, nil, head.Key())
		if err != nil {
			return err
		}
		commitedIDs := make(map[abi.SectorNumber]struct{}, len(activeSet))
		for _, info := range sset {
			// log.Debugf("info: : %+v", info)
			commitedIDs[info.SectorNumber] = struct{}{}
		}
		// log.Debugf("sset: : %+v", sset)
		log.Debugf("commitedIDs: ", len(commitedIDs))

		sort.Slice(list, func(i, j int) bool {
			return list[i] < list[j]
		})

		tw := tablewriter.New(
			tablewriter.Col("ID"),
			tablewriter.Col("State"),
			tablewriter.Col("OnChain"),
			tablewriter.Col("Active"),
			tablewriter.Col("Expiration"),
			tablewriter.Col("Deals"),
			tablewriter.Col("DealWeight"),
			tablewriter.NewLineCol("Error"),
			tablewriter.NewLineCol("EarlyExpiration"))

		fast := cctx.Bool("fast")

		for _, s := range list {
			st, err := nodeApi.SectorsStatus(ctx, s, !fast)
			if err != nil {
				tw.Write(map[string]interface{}{
					"ID":    s,
					"Error": err,
				})
				continue
			}
			log.Debugf("st: : %+v", st)

			// 	if cctx.Bool("show-removed") || st.State != api.SectorState(sealing.Removed) {
			// 		_, inSSet := commitedIDs[s]
			// 		_, inASet := activeIDs[s]

			// 		dw := .0
			// 		if st.Expiration-st.Activation > 0 {
			// 			dw = float64(big.Div(st.DealWeight, big.NewInt(int64(st.Expiration-st.Activation))).Uint64())
			// 		}

			// 		var deals int
			// 		for _, deal := range st.Deals {
			// 			if deal != 0 {
			// 				deals++
			// 			}
			// 		}

			// 		exp := st.Expiration
			// 		if st.OnTime > 0 && st.OnTime < exp {
			// 			exp = st.OnTime // Can be different when the sector was CC upgraded
			// 		}

			// 		m := map[string]interface{}{
			// 			"ID":      s,
			// 			"State":   color.New(stateOrder[sealing.SectorState(st.State)].col).Sprint(st.State),
			// 			"OnChain": yesno(inSSet),
			// 			"Active":  yesno(inASet),
			// 		}

			// 		if deals > 0 {
			// 			m["Deals"] = color.GreenString("%d", deals)
			// 		} else {
			// 			m["Deals"] = color.BlueString("CC")
			// 			if st.ToUpgrade {
			// 				m["Deals"] = color.CyanString("CC(upgrade)")
			// 			}
			// 		}

			// 		if !fast {
			// 			if !inSSet {
			// 				m["Expiration"] = "n/a"
			// 			} else {
			// 				m["Expiration"] = lcli.EpochTime(head.Height(), exp)

			// 				if !fast && deals > 0 {
			// 					m["DealWeight"] = units.BytesSize(dw)
			// 				}

			// 				if st.Early > 0 {
			// 					m["EarlyExpiration"] = color.YellowString(lcli.EpochTime(head.Height(), st.Early))
			// 				}
			// 			}
			// 		}

			// 		tw.Write(m)
			// 	}
		}

		return tw.Flush(os.Stdout)
	},
}

func yesno(b bool) string {
	if b {
		return color.GreenString("YES")
	}
	return color.RedString("NO")
}
