#![no_main]
sp1_zkvm::entrypoint!(main);

use alloy_sol_types::SolType;
use win_lib::{win, PublicValuesStruct};

pub fn main() {
    let n = sp1_zkvm::io::read::<u32>();

    let result = win(n);

    let bytes = PublicValuesStruct::abi_encode(&PublicValuesStruct { n, result });

    sp1_zkvm::io::commit_slice(&bytes);
}
