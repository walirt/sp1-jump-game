use alloy_sol_types::sol;

sol! {
    /// The public values encoded as a struct that can be easily deserialized inside Solidity.
    struct PublicValuesStruct {
        uint32 n;   
        bool result;
    }
}

/// Check if the input n equals 10, returning true if it does, false otherwise.
pub fn win(n: u32) -> bool {
    n == 10
}
