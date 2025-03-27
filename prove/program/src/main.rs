#![no_main]
sp1_zkvm::entrypoint!(main);

use win_lib::win;

pub fn main() {
    let n = sp1_zkvm::io::read::<u32>();
    sp1_zkvm::io::commit(&n);
    let result = win(n);
    sp1_zkvm::io::commit(&result);
}
