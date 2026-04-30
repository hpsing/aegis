// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {ReentrancyGuard} from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import {SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {IVerifierRegistry} from "./interfaces/IVerifierRegistry.sol";
import {IUSDC} from "./interfaces/IUSDC.sol";
import {AegisMath} from "./lib/AegisMath.sol";

/// @title AegisContract
///
/// Lifecycle:
///   1. Client calls postJob(specHash, executor, reimbursement, fee, bounty)
///      Locks (reimbursement + fee + bounty) USDC in escrow.
///   2. Executor performs the on-chain work using THEIR OWN capital,
///      with the swap's `recipient` set to the client (rug-pull guard).
///      Then calls submitClaim(jobId, txHash, reportedOutcomeHash).
///   3. Verifiers commit hashed votes during the commit window.
///   4. Verifiers reveal during the reveal window.
///   5. Anyone calls settle() once the reveal deadline has passed.
///
/// Settlement money rules:
///   - Bounty: split equally among MAJORITY revealers regardless of verdict.
///   - PASS: executor receives reimbursement + fee.
///   - FAIL: client refunded reimbursement + fee. Executor (if registered)
///     slashed SLASH_EXECUTOR_ON_FAIL: 50% to client, 50% split among
///     majority revealers.
///   - Verifier minority dissent: each slashed SLASH_PER_DISSENT to treasury.
contract AegisContract is ReentrancyGuard, Ownable {
    using SafeERC20 for IERC20;

    // uint64 public constant DEFAULT_COMMIT_WINDOW = 5 minutes;
    // uint64 public constant DEFAULT_REVEAL_WINDOW = 5 minutes;
    // uint64 public constant DEFAULT_CLAIM_WINDOW = 30 minutes;
    // Short windows for hackathon only.
    uint64 public constant DEFAULT_COMMIT_WINDOW = 30 seconds; 
    uint64 public constant DEFAULT_REVEAL_WINDOW = 30 seconds;
    uint64 public constant DEFAULT_CLAIM_WINDOW = 2 minutes;

    /// @notice Minimum reveals required to settle, censoring (verifier commits but doesn't reveal) must not
    /// deadlock the system — we proceed with as few as 2 reveals out of 3.
    /// Keeps a single rogue verifier from rubber-stamping their own job
    /// while tolerating one censorer.
    uint256 public constant MIN_REVEALS = 2;

    /// @notice Slash applied per minority verifier on settle. 10% of the
    /// 100 USDC verifier minimum stake. Routed to the treasury.
    uint256 public constant SLASH_PER_DISSENT = 10e6;

    /// @notice Slash applied to the executor on FAIL. 25% of the 500 USDC
    /// executor bond. Split 50/50:
    /// client (compensation) and majority revealers (bonus).
    uint256 public constant SLASH_EXECUTOR_ON_FAIL = 125e6;


    enum Status {
        None, // never created (default)
        Posted, // postJob done, awaiting submitClaim
        ClaimSubmitted, // commit phase open
        Settled, // settle() done; escrow flowed
        Cancelled // executor missed claim window; client refunded
    }

    /// @notice The escrow is decomposed into THREE line items per
    /// docs/architecture.md §6.2. Settled independently in settle().
    struct Job {
        address client;
        address executor;
        uint256 executorReimbursement; // refunds executor's working capital on PASS
        uint256 executorFee; // service fee paid to executor on PASS
        uint256 verifierBounty; // paid to majority revealers regardless of verdict
        bytes32 specHash; // keccak256 of canonical-JSON spec (computed off-chain)
        bytes32 txHash;
        bytes32 reportedOutcomeHash;
        uint64 claimDeadline;
        uint64 commitDeadline;
        uint64 revealDeadline;
        Status status;
    }

    struct Reveal {
        bool verdict; // true = PASS
        bool revealed; // distinguishes "voted false" from "never revealed"
    }

    IUSDC public immutable usdc;
    IVerifierRegistry public immutable registry;

    /// @notice Receives slashed minority-verifier funds. Set in constructor.
    /// In v2 this address would distribute to honest long-term verifiers
    /// or burn — out of hackathon scope.
    address public immutable treasury;

    uint256 public nextJobId = 1;
    mapping(uint256 => Job) public jobs;
    mapping(uint256 => mapping(address => bytes32)) public commits;
    mapping(uint256 => mapping(address => Reveal)) public reveals;
    mapping(uint256 => address[]) public jobRevealers;


    event JobPosted(
        uint256 indexed jobId,
        address indexed client,
        address indexed executor,
        bytes32 specHash,
        uint256 executorReimbursement,
        uint256 executorFee,
        uint256 verifierBounty
    );
    event ClaimSubmitted(
        uint256 indexed jobId,
        bytes32 txHash,
        bytes32 reportedOutcomeHash
    );
    event VoteCommitted(
        uint256 indexed jobId,
        address indexed verifier,
        bytes32 commitHash
    );
    event VoteRevealed(
        uint256 indexed jobId,
        address indexed verifier,
        bool verdict
    );
    event JobSettled(
        uint256 indexed jobId,
        bool finalVerdict,
        uint256 forVotes,
        uint256 totalReveals
    );
    event JobCancelled(uint256 indexed jobId);


    error JobNotInStatus(uint256 jobId, Status expected, Status actual);
    error DeadlineNotPassed(uint64 deadline, uint64 nowTs);
    error DeadlinePassed(uint64 deadline, uint64 nowTs);
    error ZeroAmount();
    error ZeroAddress();
    error AlreadyCommitted();
    error AlreadyRevealed();
    error NoCommit();
    error CommitMismatch(bytes32 expected, bytes32 actual);
    error NotExecutor();
    error NotClient();
    error VerifierNotActive();
    error TooFewReveals(uint256 got, uint256 required);


    constructor(
        IUSDC _usdc,
        IVerifierRegistry _registry,
        address _treasury,
        address initialOwner
    ) Ownable(initialOwner) {
        if (_treasury == address(0)) revert ZeroAddress();
        usdc = _usdc;
        registry = _registry;
        treasury = _treasury;
    }


    /// @notice Post a job. Pulls (reimbursement + fee + bounty) USDC from
    /// msg.sender and locks it in escrow. `specHash` MUST be keccak256
    /// of the canonical-JSON spec — see internal/envelope/canonical.go
    /// on the agent side, anchored by contracts/test/CanonicalDigest.t.sol.
    function postJob(
        bytes32 specHash,
        address executor,
        uint256 executorReimbursement,
        uint256 executorFee,
        uint256 verifierBounty
    ) external nonReentrant returns (uint256 jobId) {
        // Reimbursement and bounty must be non-zero (a job with no executor
        // refund makes no sense; a job with no verifier bounty has no
        // verifiers showing up). The fee CAN be zero — useful for
        // trustless execution where the client is paying the executor
        // entirely via slashing-pool exposure.
        if (executorReimbursement == 0 || verifierBounty == 0)
            revert ZeroAmount();

        uint256 total = executorReimbursement + executorFee + verifierBounty;
        IERC20(address(usdc)).safeTransferFrom(
            msg.sender,
            address(this),
            total
        );

        jobId = nextJobId++;
        jobs[jobId] = Job({
            client: msg.sender,
            executor: executor,
            executorReimbursement: executorReimbursement,
            executorFee: executorFee,
            verifierBounty: verifierBounty,
            specHash: specHash,
            txHash: bytes32(0),
            reportedOutcomeHash: bytes32(0),
            claimDeadline: uint64(block.timestamp) + DEFAULT_CLAIM_WINDOW,
            commitDeadline: 0,
            revealDeadline: 0,
            status: Status.Posted
        });

        emit JobPosted(
            jobId,
            msg.sender,
            executor,
            specHash,
            executorReimbursement,
            executorFee,
            verifierBounty
        );
    }

    /// @notice Executor declares the claim. Locks in commit/reveal windows.
    /// The contract does NOT validate the on-chain claim itself — that's
    /// the verifier swarm's job.
    function submitClaim(
        uint256 jobId,
        bytes32 txHash,
        bytes32 reportedOutcomeHash
    ) external {
        Job storage j = jobs[jobId];
        if (j.status != Status.Posted) {
            revert JobNotInStatus(jobId, Status.Posted, j.status);
        }
        if (msg.sender != j.executor) revert NotExecutor();
        if (uint64(block.timestamp) > j.claimDeadline) {
            revert DeadlinePassed(j.claimDeadline, uint64(block.timestamp));
        }

        j.txHash = txHash;
        j.reportedOutcomeHash = reportedOutcomeHash;
        j.commitDeadline = uint64(block.timestamp) + DEFAULT_COMMIT_WINDOW;
        j.revealDeadline = j.commitDeadline + DEFAULT_REVEAL_WINDOW;
        j.status = Status.ClaimSubmitted;

        emit ClaimSubmitted(jobId, txHash, reportedOutcomeHash);
    }

    /// @notice Commit a hashed vote. commitHash MUST equal
    ///   keccak256(abi.encode(verdict, nonce, msg.sender))
    /// Including msg.sender prevents copy-and-front-run by another verifier.
    function commitVote(uint256 jobId, bytes32 commitHash) external {
        Job storage j = jobs[jobId];
        if (j.status != Status.ClaimSubmitted) {
            revert JobNotInStatus(jobId, Status.ClaimSubmitted, j.status);
        }
        if (uint64(block.timestamp) > j.commitDeadline) {
            revert DeadlinePassed(j.commitDeadline, uint64(block.timestamp));
        }
        if (!registry.isActive(msg.sender)) revert VerifierNotActive();
        if (commits[jobId][msg.sender] != bytes32(0)) revert AlreadyCommitted();

        commits[jobId][msg.sender] = commitHash;
        emit VoteCommitted(jobId, msg.sender, commitHash);
    }

    /// @notice Reveal a previously committed vote. The reveal window opens
    /// AFTER commit deadline (prevents observing other reveals before
    /// committing).
    function revealVote(uint256 jobId, bool verdict, bytes32 nonce) external {
        Job storage j = jobs[jobId];
        if (j.status != Status.ClaimSubmitted) {
            revert JobNotInStatus(jobId, Status.ClaimSubmitted, j.status);
        }
        if (uint64(block.timestamp) <= j.commitDeadline) {
            revert DeadlineNotPassed(j.commitDeadline, uint64(block.timestamp));
        }
        if (uint64(block.timestamp) > j.revealDeadline) {
            revert DeadlinePassed(j.revealDeadline, uint64(block.timestamp));
        }
        bytes32 expected = commits[jobId][msg.sender];
        if (expected == bytes32(0)) revert NoCommit();
        if (reveals[jobId][msg.sender].revealed) revert AlreadyRevealed();

        bytes32 actual = keccak256(abi.encode(verdict, nonce, msg.sender));
        if (actual != expected) revert CommitMismatch(expected, actual);

        reveals[jobId][msg.sender] = Reveal({verdict: verdict, revealed: true});
        jobRevealers[jobId].push(msg.sender);

        emit VoteRevealed(jobId, msg.sender, verdict);
    }

    /// @notice Tally reveals, distribute bounty, run executor settlement,
    /// slash dissenters.
    /// Permissionless; first caller wins, others revert via the status guard.
    function settle(uint256 jobId) external nonReentrant {
        Job storage j = jobs[jobId];
        if (j.status != Status.ClaimSubmitted) {
            revert JobNotInStatus(jobId, Status.ClaimSubmitted, j.status);
        }
        if (uint64(block.timestamp) <= j.revealDeadline) {
            revert DeadlineNotPassed(j.revealDeadline, uint64(block.timestamp));
        }

        address[] storage revs = jobRevealers[jobId];
        uint256 totalReveals = revs.length;
        if (totalReveals < MIN_REVEALS)
            revert TooFewReveals(totalReveals, MIN_REVEALS);

        uint256 forVotes;
        for (uint256 i = 0; i < totalReveals; i++) {
            if (reveals[jobId][revs[i]].verdict) forVotes++;
        }
        bool finalVerdict = AegisMath.is2of3Majority(forVotes, totalReveals);

        // Mark settled BEFORE any external interaction.
        j.status = Status.Settled;

        uint256 majorityCount = finalVerdict
            ? forVotes
            : (totalReveals - forVotes);
        uint256 perVerifierBounty = majorityCount > 0
            ? j.verifierBounty / majorityCount
            : 0;

        // Pay/slash verifiers based on majority alignment.
        for (uint256 i = 0; i < totalReveals; i++) {
            address v = revs[i];
            bool wasCorrect = reveals[jobId][v].verdict == finalVerdict;
            registry.recordVote(v, wasCorrect);
            if (wasCorrect) {
                if (perVerifierBounty > 0) {
                    IERC20(address(usdc)).safeTransfer(v, perVerifierBounty);
                }
            } else {
                registry.slash(v, SLASH_PER_DISSENT, treasury);
            }
        }

        // Settle the executor side.
        uint256 reimbursementPlusFee = j.executorReimbursement + j.executorFee;
        if (finalVerdict) {
            // PASS: executor gets back working capital + earns fee.
            IERC20(address(usdc)).safeTransfer(
                j.executor,
                reimbursementPlusFee
            );
        } else {
            // FAIL: client refunded both line items; executor slashed if
            // they're registered (bonded). Slash splits 50/50 between
            // client (compensation) and majority revealers (bonus).
            IERC20(address(usdc)).safeTransfer(j.client, reimbursementPlusFee);
            if (registry.stakeOf(j.executor) > 0) {
                _slashExecutor(
                    jobId,
                    j.executor,
                    j.client,
                    revs,
                    totalReveals,
                    finalVerdict
                );
            }
        }

        emit JobSettled(jobId, finalVerdict, forVotes, totalReveals);
    }

    /// @dev Splits SLASH_EXECUTOR_ON_FAIL 50/50 between client and majority.
    /// Each call to registry.slash is bounded by the executor's remaining stake — if stake runs out
    /// mid-distribution, later calls slash less or zero. Acceptable
    /// degradation: the system never tries to slash more than exists.
    function _slashExecutor(
        uint256 jobId,
        address executor,
        address client,
        address[] storage revs,
        uint256 totalReveals,
        bool finalVerdict
    ) private {
        uint256 toClient = SLASH_EXECUTOR_ON_FAIL / 2;
        uint256 toVerifiers = SLASH_EXECUTOR_ON_FAIL - toClient;

        registry.slash(executor, toClient, client);

        uint256 majorityCount;
        for (uint256 i = 0; i < totalReveals; i++) {
            if (reveals[jobId][revs[i]].verdict == finalVerdict)
                majorityCount++;
        }
        if (majorityCount == 0) return;

        uint256 perVerifier = toVerifiers / majorityCount;
        if (perVerifier == 0) return;

        for (uint256 i = 0; i < totalReveals; i++) {
            address v = revs[i];
            if (reveals[jobId][v].verdict == finalVerdict) {
                registry.slash(executor, perVerifier, v);
            }
        }
    }

    /// @notice If the executor never submitted a claim, the client can
    /// recover the full escrow after the claim deadline passes.
    function cancelStaleJob(uint256 jobId) external nonReentrant {
        Job storage j = jobs[jobId];
        if (j.status != Status.Posted) {
            revert JobNotInStatus(jobId, Status.Posted, j.status);
        }
        if (uint64(block.timestamp) <= j.claimDeadline) {
            revert DeadlineNotPassed(j.claimDeadline, uint64(block.timestamp));
        }
        if (msg.sender != j.client) revert NotClient();

        j.status = Status.Cancelled;
        uint256 total = j.executorReimbursement +
            j.executorFee +
            j.verifierBounty;
        IERC20(address(usdc)).safeTransfer(j.client, total);
        emit JobCancelled(jobId);
    }

    function revealCount(uint256 jobId) external view returns (uint256) {
        return jobRevealers[jobId].length;
    }
}
