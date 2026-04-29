// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { AegisContract } from "../src/AegisContract.sol";
import { TestSetup } from "./helpers/TestSetup.sol";

/// @notice Revert-path coverage for AegisContract — every branch
/// AegisContractTest doesn't already exercise.
contract AegisContractEdges is TestSetup {
    bytes32 internal constant SPEC_HASH = keccak256("spec");

    function test_PostJob_ZeroReimbursementReverts() public {
        vm.prank(client);
        vm.expectRevert(AegisContract.ZeroAmount.selector);
        aegis.postJob(SPEC_HASH, executor, 0, FEE, BOUNTY);
    }

    function test_PostJob_ZeroBountyReverts() public {
        vm.prank(client);
        vm.expectRevert(AegisContract.ZeroAmount.selector);
        aegis.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, 0);
    }

    function test_PostJob_ZeroFeeAccepted() public {
        // A zero-fee job is legitimate (purely altruistic verification).
        vm.prank(client);
        uint256 jobId = aegis.postJob(SPEC_HASH, executor, REIMBURSEMENT, 0, BOUNTY);
        assertGt(jobId, 0);
    }

    function test_SubmitClaim_NonExecutorReverts() public {
        vm.prank(client);
        uint256 jobId = aegis.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, BOUNTY);
        vm.prank(client); // wrong actor
        vm.expectRevert(AegisContract.NotExecutor.selector);
        aegis.submitClaim(jobId, keccak256("tx"), keccak256("outcome"));
    }

    function test_SubmitClaim_AfterClaimDeadlineReverts() public {
        vm.prank(client);
        uint256 jobId = aegis.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, BOUNTY);
        vm.warp(block.timestamp + aegis.DEFAULT_CLAIM_WINDOW() + 1);
        vm.prank(executor);
        vm.expectRevert();
        aegis.submitClaim(jobId, keccak256("tx"), keccak256("outcome"));
    }

    function test_SubmitClaim_TwiceReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        vm.prank(executor);
        vm.expectRevert();
        aegis.submitClaim(jobId, keccak256("tx2"), keccak256("outcome2"));
    }

    function test_CommitVote_AlreadyCommittedReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 h = _commitHash(true, keccak256("n"), verifier1);
        vm.prank(verifier1);
        aegis.commitVote(jobId, h);
        vm.prank(verifier1);
        vm.expectRevert(AegisContract.AlreadyCommitted.selector);
        aegis.commitVote(jobId, h);
    }

    function test_CommitVote_NotActiveReverts() public {
        // Use an unregistered address.
        address randomCaller = address(0xDEAD);
        uint256 jobId = _postAndClaim(SPEC_HASH);
        vm.prank(randomCaller);
        vm.expectRevert(AegisContract.VerifierNotActive.selector);
        aegis.commitVote(jobId, _commitHash(true, keccak256("n"), randomCaller));
    }

    function test_RevealVote_BeforeCommitDeadlineReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n = keccak256("n");
        _commitAndReveal(jobId, verifier1, true, n);
        // Don't warp — reveal phase isn't open yet.
        vm.prank(verifier1);
        vm.expectRevert();
        aegis.revealVote(jobId, true, n);
    }

    function test_RevealVote_NoCommitReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        vm.warp(block.timestamp + aegis.DEFAULT_COMMIT_WINDOW() + 1);
        vm.prank(verifier1);
        vm.expectRevert(AegisContract.NoCommit.selector);
        aegis.revealVote(jobId, true, keccak256("n"));
    }

    function test_RevealVote_AlreadyRevealedReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n = keccak256("n");
        _commitAndReveal(jobId, verifier1, true, n);
        vm.warp(block.timestamp + aegis.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n);
        vm.prank(verifier1);
        vm.expectRevert(AegisContract.AlreadyRevealed.selector);
        aegis.revealVote(jobId, true, n);
    }

    function test_Settle_BeforeRevealDeadlineReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        _commitAndReveal(jobId, verifier1, true, n1);
        _commitAndReveal(jobId, verifier2, true, n2);
        _commitAndReveal(jobId, verifier3, true, n3);
        vm.warp(block.timestamp + aegis.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n1);
        _doReveal(jobId, verifier2, true, n2);
        _doReveal(jobId, verifier3, true, n3);
        // Don't warp past reveal deadline.
        vm.expectRevert();
        aegis.settle(jobId);
    }

    function test_Settle_TooFewRevealsReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n = keccak256("n");
        _commitAndReveal(jobId, verifier1, true, n);
        vm.warp(block.timestamp + aegis.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n);
        vm.warp(block.timestamp + aegis.DEFAULT_REVEAL_WINDOW() + 1);
        vm.expectRevert();
        aegis.settle(jobId);
    }

    function test_CancelStaleJob_BeforeDeadlineReverts() public {
        vm.prank(client);
        uint256 jobId = aegis.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, BOUNTY);
        vm.prank(client);
        vm.expectRevert();
        aegis.cancelStaleJob(jobId);
    }

    function test_CancelStaleJob_NonClientReverts() public {
        vm.prank(client);
        uint256 jobId = aegis.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, BOUNTY);
        vm.warp(block.timestamp + aegis.DEFAULT_CLAIM_WINDOW() + 1);
        vm.prank(executor);
        vm.expectRevert(AegisContract.NotClient.selector);
        aegis.cancelStaleJob(jobId);
    }

    function test_RevealCount_View() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n = keccak256("n");
        _commitAndReveal(jobId, verifier1, true, n);
        vm.warp(block.timestamp + aegis.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n);
        assertEq(aegis.revealCount(jobId), 1);
    }
}
