## sector-sanity-checker-v0.3.0

This tools can help you check sector to avoid the window PoST fail.

## Download

https://github.com/irocn/sector-checker/releases/tag/sector-sanity-checker-v0.3.0

## Usage

```bash
├── api
├── cache
│   └── s-t000100-156
│       ├── p_aux
│       ├── sc-02-data-tree-r-last-0.dat
│       ├── sc-02-data-tree-r-last-1.dat
│       ├── sc-02-data-tree-r-last-2.dat
│       ├── sc-02-data-tree-r-last-3.dat
│       ├── sc-02-data-tree-r-last-4.dat
│       ├── sc-02-data-tree-r-last-5.dat
│       ├── sc-02-data-tree-r-last-6.dat
│       ├── sc-02-data-tree-r-last-7.dat
│       └── t_aux
├── config.toml
├── sealed
│   └── s-t000100-156
├── token
└── unsealed

```

### step 1, export the environment variable
 - export FIL_PROOFS_PARENT_CACHE=<YOUR_PARENT_CACHE>
 - export FIL_PROOFS_PARAMETER_CACHE=<YOUR_FIL_PROOFS_PARAMETER_CACHE>
 - export FIL_PROOFS_USE_GPU_COLUMN_BUILDER=1 
 - export RUST_LOG=info FIL_PROOFS_USE_GPU_TREE_BUILDER=1 
 - export FIL_PROOFS_MAXIMIZE_CACHING=1
 - export MINER_API_INFO=<YOUR_MINER_API_INFO>
### step 2, run the tool 
 - $>sector-sanity-checker checking  --sector-size=32G --miner-addr=<your_miner_id> --storage-dir=<sector_dir> 
 - $>sector-sanity-checker checking  --sector-size=32G --sectors-file-only-number=<sectors-to-scan> --miner-addr=<your_miner_id> --storage-dir=<sector_dir>
 
### For Example:

 - $>sector-sanity-checker checking  --sector-size=32G --miner-addr=t### --storage-dir=/opt/data/storage
 
 Then all the sectors under /opt/data/storage/sealed/s-xxxxx-xxx will be scaned.
 
 - $>sector-sanity-checker checking  --sector-size=32G --sectors-file-only-number=sectors-to-scan.txt --miner-addr=t### --storage-dir=/opt/data/storage
 

## donate
If the tool help you, please donate 1 FIL to us.
 - Wallet Address: f3rozq5jglh7m42v37gsij477kpwnpuamlr5p7daf7t54szc6nbv4kzi7wgw4ea2apxjlopi5nwdbvccvxnymq
