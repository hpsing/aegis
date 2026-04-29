// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

/// @notice Test stand-in for USDC. 6 decimals like the real thing.
/// Free mint via `mint()` so tests can hand verifiers stake.
contract MockUSDC is ERC20 {
    constructor() ERC20("MockUSDC", "mUSDC") { }

    function decimals() public pure override returns (uint8) {
        return 6;
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }
}
