// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Test } from "forge-std/Test.sol";

import { AegisContract } from "../../src/AegisContract.sol";
import { VerifierRegistry } from "../../src/VerifierRegistry.sol";
import { IUSDC } from "../../src/interfaces/IUSDC.sol";
import { MockUSDC } from "../mocks/MockUSDC.sol";

/// @notice Shared deployment + actor harness. Inherit and call setUp().
///   - REIMBURSEMENT — refunds executor working capital on PASS
///   - FEE           — service fee paid to executor on PASS
///   - BOUNTY        — paid to majority verifiers regardless of verdict
abstract contract TestSetup is Test {
    MockUSDC internal usdc;
    VerifierRegistry internal registry;
    AegisContract internal aegis;

    address internal owner = address(0xA1);
    address internal treasury = address(0x77);
    address internal client = address(0xC1);
    address internal executor = address(0xE1);
    address internal verifier1 = address(0xB1);
    address internal verifier2 = address(0xB2);
    address internal verifier3 = address(0xB3);

    uint256 internal constant VERIFIER_STAKE = 1000e6; // 1,000 USDC bond
    uint256 internal constant EXECUTOR_STAKE = 1000e6; // tests register executor too
    uint256 internal constant REIMBURSEMENT = 100e6; // executor working capital refund
    uint256 internal constant FEE = 10e6; // service fee
    uint256 internal constant BOUNTY = 30e6; // verifier bounty (10/10/10 split)

    function setUp() public virtual {
        usdc = new MockUSDC();

        vm.startPrank(owner);
        registry = new VerifierRegistry(IUSDC(address(usdc)), owner);
        aegis = new AegisContract(IUSDC(address(usdc)), registry, treasury, owner);
        registry.setAegis(address(aegis));
        vm.stopPrank();

        // Fund actors generously.
        usdc.mint(client, 100_000e6);
        usdc.mint(executor, 100_000e6);
        for (uint256 i = 0; i < 3; i++) {
            usdc.mint(_verifier(i), 100_000e6);
        }

        // Register all 3 verifiers.
        for (uint256 i = 0; i < 3; i++) {
            address v = _verifier(i);
            vm.startPrank(v);
            usdc.approve(address(registry), VERIFIER_STAKE);
            registry.register(VERIFIER_STAKE, 0);
            vm.stopPrank();
        }

        // Client approves aegis to pull (reimbursement + fee + bounty).
        vm.prank(client);
        usdc.approve(address(aegis), type(uint256).max);
    }

    function _verifier(
        uint256 i
    ) internal view returns (address) {
        if (i == 0) return verifier1;
        if (i == 1) return verifier2;
        return verifier3;
    }

    /// @notice Register the executor too (needed for FAIL-slash tests).
    /// Many tests don't need this — only the ones exercising the slash path.
    function _registerExecutor() internal {
        vm.startPrank(executor);
        usdc.approve(address(registry), EXECUTOR_STAKE);
        registry.register(EXECUTOR_STAKE, 0);
        vm.stopPrank();
    }

    /// @notice Post a job + submit a claim, returning the jobId.
    function _postAndClaim(
        bytes32 specHash
    ) internal returns (uint256 jobId) {
        vm.prank(client);
        jobId = aegis.postJob(specHash, executor, REIMBURSEMENT, FEE, BOUNTY);
        vm.prank(executor);
        aegis.submitClaim(jobId, keccak256("tx"), keccak256("outcome"));
    }

    function _commitHash(
        bool verdict,
        bytes32 nonce,
        address verifier
    ) internal pure returns (bytes32) {
        return keccak256(abi.encode(verdict, nonce, verifier));
    }

    function _commitAndReveal(
        uint256 jobId,
        address verifier,
        bool verdict,
        bytes32 nonce
    ) internal {
        vm.prank(verifier);
        aegis.commitVote(jobId, _commitHash(verdict, nonce, verifier));
    }

    function _doReveal(uint256 jobId, address verifier, bool verdict, bytes32 nonce) internal {
        vm.prank(verifier);
        aegis.revealVote(jobId, verdict, nonce);
    }
}
