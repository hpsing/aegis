// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

/// @notice The minimum surface AegisContract needs from the registry.
/// Decoupled so AegisContract can be tested with mock registries.
///
/// The registry holds bonded participants — both verifiers (100 USDC min)
/// and executors (500 USDC bond) live here. They differ only by stake size
/// and operational role; on-chain they're treated uniformly. (Naming is
/// historical; see docs/architecture.md §8.2.)
interface IVerifierRegistry {
    /// @notice Slash `verifier` by up to `amount` USDC and TRANSFER the
    /// slashed amount out of the registry to `recipient`. Returns the
    /// actual amount slashed (bounded by the staker's current balance,
    /// never underflows). Only callable by AegisContract.
    ///
    /// Recipient routing per docs/architecture.md §6.5–§6.6:
    ///   - verifier minority dissent  -> treasury
    ///   - executor FAIL slash, half  -> client (compensation)
    ///   - executor FAIL slash, other half -> majority verifiers (split)
    function slash(
        address verifier,
        uint256 amount,
        address recipient
    ) external returns (uint256 actualSlashed);

    /// @notice Record a single vote outcome on a verifier's lifetime ledger.
    /// Only callable by AegisContract.
    function recordVote(address verifier, bool wasCorrect) external;

    /// @notice True iff `verifier` is currently registered, active, and
    /// holding at least minStake. AegisContract uses this to filter
    /// commits from non-registered identities.
    function isActive(
        address verifier
    ) external view returns (bool);

    /// @notice Stake currently bonded by `verifier` (USDC, 6 decimals).
    function stakeOf(
        address verifier
    ) external view returns (uint256);
}
