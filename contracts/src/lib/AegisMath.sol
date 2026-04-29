// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

/// @notice Pure-integer math for Aegis's settlement logic.
/// Avoids floating-point so the contract and the verifier agents
/// agree on the threshold byte-for-byte.
library AegisMath {
    /// @notice Returns true iff `forVotes >= ceil(2/3 * total)`.
    /// Equivalent to `forVotes * 3 >= total * 2` for integers — avoids
    /// any divide-and-round subtlety.
    function is2of3Majority(uint256 forVotes, uint256 total) internal pure returns (bool) {
        if (total == 0) return false;
        return forVotes * 3 >= total * 2;
    }

    /// @notice Bps accuracy = correct * 10_000 / total, capped at 10_000.
    /// Returns 0 when total is 0 (no votes cast yet).
    function accuracyBps(uint256 correct, uint256 total) internal pure returns (uint256) {
        if (total == 0) return 0;
        uint256 bps = (correct * 10_000) / total;
        return bps > 10_000 ? 10_000 : bps;
    }
}
