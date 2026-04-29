// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

/// @notice We talk to USDC via the standard IERC20 surface; this alias
/// just makes the intent explicit at call sites. Actual USDC has 6
/// decimals — callers must scale amounts accordingly (constants like
/// MIN_STAKE = 100e6 are correct for 100 USDC).
interface IUSDC is IERC20 { }
