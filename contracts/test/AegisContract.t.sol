// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { AegisContract } from "../src/AegisContract.sol";
import { VerifierRegistry } from "../src/VerifierRegistry.sol";
import { IUSDC } from "../src/interfaces/IUSDC.sol";
import { ReentrantUSDC } from "./mocks/ReentrantUSDC.sol";
import { TestSetup } from "./helpers/TestSetup.sol";

contract AegisContractTest is TestSetup {
    bytes32 internal constant SPEC_HASH = keccak256("spec");

    function test_HappyPath_3VerifiersAllAgreePass() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        uint256 executorBalBefore = usdc.balanceOf(executor);

        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        _commitAndReveal(jobId, verifier1, true, n1);
        _commitAndReveal(jobId, verifier2, true, n2);
        _commitAndReveal(jobId, verifier3, true, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n1);
        _doReveal(jobId, verifier2, true, n2);
        _doReveal(jobId, verifier3, true, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_REVEAL_WINDOW() + 1);
        quorum.settle(jobId);

        // PASS: executor receives reimbursement + fee.
        assertEq(
            usdc.balanceOf(executor),
            executorBalBefore + REIMBURSEMENT + FEE,
            "executor receives reimbursement + fee"
        );
        // No verifier slashed, all in majority.
        for (uint256 i = 0; i < 3; i++) {
            assertEq(registry.stakeOf(_verifier(i)), VERIFIER_STAKE, "no slash on majority side");
        }
    }

    function test_HappyPath_3VerifiersAllAgreeFail() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        uint256 clientBalBefore = usdc.balanceOf(client);

        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        _commitAndReveal(jobId, verifier1, false, n1);
        _commitAndReveal(jobId, verifier2, false, n2);
        _commitAndReveal(jobId, verifier3, false, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, false, n1);
        _doReveal(jobId, verifier2, false, n2);
        _doReveal(jobId, verifier3, false, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_REVEAL_WINDOW() + 1);
        quorum.settle(jobId);

        // FAIL with executor not registered: client gets reimbursement + fee back,
        // but no executor slash flows (best-effort).
        assertEq(
            usdc.balanceOf(client),
            clientBalBefore + REIMBURSEMENT + FEE,
            "client refunded reimbursement + fee"
        );
    }

    /// @notice The full FAIL flow: executor IS registered, so the slash
    /// fires and splits 50/50 between client compensation and the
    /// majority verifiers (all 3 voted FAIL → all 3 get a share).
    function test_FailScenario_ExecutorRegistered_SlashSplits50_50() public {
        _registerExecutor();
        uint256 jobId = _postAndClaim(SPEC_HASH);

        uint256 clientBalBefore = usdc.balanceOf(client);
        uint256[3] memory verifierBalsBefore;
        for (uint256 i = 0; i < 3; i++) {
            verifierBalsBefore[i] = usdc.balanceOf(_verifier(i));
        }

        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        _commitAndReveal(jobId, verifier1, false, n1);
        _commitAndReveal(jobId, verifier2, false, n2);
        _commitAndReveal(jobId, verifier3, false, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, false, n1);
        _doReveal(jobId, verifier2, false, n2);
        _doReveal(jobId, verifier3, false, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_REVEAL_WINDOW() + 1);
        quorum.settle(jobId);

        uint256 slashTotal = quorum.SLASH_EXECUTOR_ON_FAIL();
        uint256 toClient = slashTotal / 2;
        uint256 toVerifiers = slashTotal - toClient;
        uint256 perVerifier = toVerifiers / 3;
        uint256 bountyShare = BOUNTY / 3;

        // Client: reimbursement + fee refund + half the slash.
        assertEq(
            usdc.balanceOf(client),
            clientBalBefore + REIMBURSEMENT + FEE + toClient,
            "client got refund + 50% slash compensation"
        );
        // Each majority verifier: bounty share + 1/3 of the verifier-side slash.
        for (uint256 i = 0; i < 3; i++) {
            assertEq(
                usdc.balanceOf(_verifier(i)),
                verifierBalsBefore[i] + bountyShare + perVerifier,
                "verifier got bounty + slash share"
            );
        }
        // Executor stake reduced by the actual slashed total
        // (perVerifier dust may leave a tiny remainder).
        uint256 expectedExecutorStakeAfter = EXECUTOR_STAKE - toClient - (perVerifier * 3);
        assertEq(registry.stakeOf(executor), expectedExecutorStakeAfter, "executor slashed");
    }

    function test_2of3MajorityPass() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);

        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        _commitAndReveal(jobId, verifier1, true, n1);
        _commitAndReveal(jobId, verifier2, true, n2);
        _commitAndReveal(jobId, verifier3, false, n3); // dissenter

        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n1);
        _doReveal(jobId, verifier2, true, n2);
        _doReveal(jobId, verifier3, false, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_REVEAL_WINDOW() + 1);
        uint256 treasuryBalBefore = usdc.balanceOf(treasury);
        quorum.settle(jobId);

        // V3 dissented → slashed by SLASH_PER_DISSENT, routed to treasury.
        assertEq(
            registry.stakeOf(verifier3),
            VERIFIER_STAKE - quorum.SLASH_PER_DISSENT(),
            "dissenter slashed"
        );
        assertEq(registry.stakeOf(verifier1), VERIFIER_STAKE, "majority verifier 1 not slashed");
        assertEq(registry.stakeOf(verifier2), VERIFIER_STAKE, "majority verifier 2 not slashed");
        assertEq(
            usdc.balanceOf(treasury),
            treasuryBalBefore + quorum.SLASH_PER_DISSENT(),
            "treasury received the slashed funds"
        );
    }

    function test_RevealMustMatchCommit() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n = keccak256("n");

        _commitAndReveal(jobId, verifier1, true, n);
        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        vm.prank(verifier1);
        vm.expectRevert();
        quorum.revealVote(jobId, false, n);
    }

    function test_CannotCommitAfterDeadline() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        vm.prank(verifier1);
        vm.expectRevert();
        quorum.commitVote(jobId, _commitHash(true, keccak256("n"), verifier1));
    }

    function test_CannotRevealAfterDeadline() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n = keccak256("n");
        _commitAndReveal(jobId, verifier1, true, n);

        vm.warp(
            block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + quorum.DEFAULT_REVEAL_WINDOW() + 1
        );
        vm.prank(verifier1);
        vm.expectRevert();
        quorum.revealVote(jobId, true, n);
    }

    function test_DoubleSettleReverts() public {
        uint256 jobId = _postAndClaim(SPEC_HASH);
        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        _commitAndReveal(jobId, verifier1, true, n1);
        _commitAndReveal(jobId, verifier2, true, n2);
        _commitAndReveal(jobId, verifier3, true, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_COMMIT_WINDOW() + 1);
        _doReveal(jobId, verifier1, true, n1);
        _doReveal(jobId, verifier2, true, n2);
        _doReveal(jobId, verifier3, true, n3);

        vm.warp(block.timestamp + quorum.DEFAULT_REVEAL_WINDOW() + 1);
        quorum.settle(jobId);

        vm.expectRevert();
        quorum.settle(jobId);
    }

    function test_CancelStaleJobRefundsFullEscrow() public {
        vm.prank(client);
        uint256 jobId = quorum.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, BOUNTY);

        vm.warp(block.timestamp + quorum.DEFAULT_CLAIM_WINDOW() + 1);

        uint256 clientBalBefore = usdc.balanceOf(client);
        vm.prank(client);
        quorum.cancelStaleJob(jobId);
        assertEq(
            usdc.balanceOf(client),
            clientBalBefore + REIMBURSEMENT + FEE + BOUNTY,
            "full 3-line-item refund"
        );
    }

    function test_SlashingCannotUnderflow() public {
        // Drive verifier3's stake to zero by repeated quorum-driven slashes.
        vm.startPrank(address(quorum));
        for (uint256 i = 0; i < 200; i++) {
            registry.slash(verifier3, 100e6, treasury);
        }
        vm.stopPrank();
        assertEq(registry.stakeOf(verifier3), 0);
    }

    /// @notice Reentrancy: USDC transfer to executor on settle re-enters
    /// settle(). The OZ ReentrancyGuard must abort the inner call.
    function test_ReentrancyOnSettle() public {
        ReentrantUSDC rusdc = new ReentrantUSDC();
        VerifierRegistry reg2 = new VerifierRegistry(IUSDC(address(rusdc)), owner);
        AegisContract q2 = new AegisContract(IUSDC(address(rusdc)), reg2, treasury, owner);
        vm.prank(owner);
        reg2.setQuorum(address(q2));

        rusdc.mint(client, 100_000e6);
        for (uint256 i = 0; i < 3; i++) {
            address v = _verifier(i);
            rusdc.mint(v, 100_000e6);
            vm.startPrank(v);
            rusdc.approve(address(reg2), type(uint256).max);
            reg2.register(VERIFIER_STAKE, 0);
            vm.stopPrank();
        }
        vm.prank(client);
        rusdc.approve(address(q2), type(uint256).max);

        vm.prank(client);
        uint256 jobId = q2.postJob(SPEC_HASH, executor, REIMBURSEMENT, FEE, BOUNTY);
        vm.prank(executor);
        q2.submitClaim(jobId, keccak256("tx"), keccak256("outcome"));

        bytes32 n1 = keccak256("n1");
        bytes32 n2 = keccak256("n2");
        bytes32 n3 = keccak256("n3");
        vm.prank(verifier1);
        q2.commitVote(jobId, keccak256(abi.encode(true, n1, verifier1)));
        vm.prank(verifier2);
        q2.commitVote(jobId, keccak256(abi.encode(true, n2, verifier2)));
        vm.prank(verifier3);
        q2.commitVote(jobId, keccak256(abi.encode(true, n3, verifier3)));

        vm.warp(block.timestamp + q2.DEFAULT_COMMIT_WINDOW() + 1);
        vm.prank(verifier1);
        q2.revealVote(jobId, true, n1);
        vm.prank(verifier2);
        q2.revealVote(jobId, true, n2);
        vm.prank(verifier3);
        q2.revealVote(jobId, true, n3);

        rusdc.arm(executor, address(q2), jobId);

        vm.warp(block.timestamp + q2.DEFAULT_REVEAL_WINDOW() + 1);
        q2.settle(jobId);

        assertTrue(rusdc.attemptedReenter(), "the reentrant USDC must have tried");
        assertTrue(
            rusdc.reenterReverted(), "ReentrancyGuard should have blocked the recursive settle"
        );
    }
}
