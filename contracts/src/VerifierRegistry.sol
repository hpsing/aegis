// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import { IVerifierRegistry } from "./interfaces/IVerifierRegistry.sol";
import { IUSDC } from "./interfaces/IUSDC.sol";
import { AegisMath } from "./lib/AegisMath.sol";
import { VerifierINFT } from "./VerifierINFT.sol";

/// @title VerifierRegistry
/// @notice Tracks the set of staked verifiers, their stake balances,
/// their per-vote accuracy, and (set in step 6) their ERC-7857 iNFT id.
/// Slashing and vote recording are gated to the AegisContract address.
contract VerifierRegistry is IVerifierRegistry, Ownable {
    using SafeERC20 for IERC20;

    /// @notice Default minimum bonded stake required to remain active.
    /// Override via setMinStake (owner only). 100 USDC = 100 * 10^6.
    uint256 public constant DEFAULT_MIN_STAKE = 100e6;

    /// @notice Withdrawal cooldown after requestWithdraw, in blocks.
    /// TODO:: currently limiting to 7 blocks.
    uint256 public constant WITHDRAW_COOLDOWN_BLOCKS = 7;

    struct Verifier {
        uint256 stake; // current bonded amount (USDC, 6 decimals)
        uint256 votesTotal; // lifetime votes cast
        uint256 votesCorrect; // lifetime votes that matched final outcome
        uint256 iNftId; // ERC-7857 token id on 0G (set in step 6, 0 default)
        uint256 withdrawableAt; // 0 if no pending withdraw, else block number
        bool active; // false if slashed below minStake or after withdraw
    }

    IUSDC public immutable usdc;

    /// @notice Optional VerifierINFT contract. If set (via setINFT, owner
    /// only), `register(amount, iNftId)` requires the caller to be the
    /// iNFT's `controller`. If unset (zero), iNFT linking is best-effort —
    /// any tokenId is accepted without verification.
    VerifierINFT public iNftContract;

    mapping(address => Verifier) public verifiers;
    address[] public verifierAddresses;
    mapping(address => bool) private isRegistered;

    address public quorum; // the AegisContract; set once via setQuorum
    uint256 public minStake;

    event VerifierRegistered(address indexed verifier, uint256 stake, uint256 iNftId);
    event StakeAdded(address indexed verifier, uint256 amount, uint256 newStake);
    event WithdrawalRequested(address indexed verifier, uint256 unlockBlock);
    event StakeWithdrawn(address indexed verifier, uint256 amount);
    event VerifierSlashed(
        address indexed verifier, address indexed recipient, uint256 amount, uint256 newStake
    );
    event VoteRecorded(
        address indexed verifier, bool wasCorrect, uint256 votesTotal, uint256 votesCorrect
    );
    event QuorumSet(address indexed quorum);
    event MinStakeUpdated(uint256 oldMin, uint256 newMin);

    error InsufficientStake(uint256 provided, uint256 required);
    error NotRegistered();
    error AlreadyRegistered();
    error NotQuorum(address caller);
    error WithdrawNotRequested();
    error CooldownNotElapsed(uint256 unlockBlock, uint256 currentBlock);
    error ZeroRecipient();
    error NotINftController(uint256 tokenId, address claimedBy);

    constructor(IUSDC _usdc, address initialOwner) Ownable(initialOwner) {
        usdc = _usdc;
        minStake = DEFAULT_MIN_STAKE;
    }

    /// @notice Wire the AegisContract address. Callable once by owner.
    function setQuorum(
        address _quorum
    ) external onlyOwner {
        require(quorum == address(0), "quorum already set");
        require(_quorum != address(0), "zero quorum");
        quorum = _quorum;
        emit QuorumSet(_quorum);
    }

    /// @notice Wire the VerifierINFT contract. Owner-only. If left as zero
    /// (the default), the registry skips iNFT controller checks — useful
    /// for local devnets where the iNFT contract isn't deployed yet.
    /// Step 6 sets this for production deployments.
    function setINftContract(
        VerifierINFT _iNft
    ) external onlyOwner {
        iNftContract = _iNft;
    }

    function setMinStake(
        uint256 _minStake
    ) external onlyOwner {
        emit MinStakeUpdated(minStake, _minStake);
        minStake = _minStake;
    }

    /// @notice Register as a verifier with a bonded stake. Caller must
    /// have approved this contract to pull `amount` USDC. iNftId is
    /// optional (pass 0 to skip); when non-zero, the caller MUST be the
    /// iNFT's current controller.
    function register(uint256 amount, uint256 iNftId) external {
        if (isRegistered[msg.sender]) revert AlreadyRegistered();
        if (amount < minStake) revert InsufficientStake(amount, minStake);
        if (iNftId != 0 && address(iNftContract) != address(0)) {
            address ctrl = iNftContract.controllerOf(iNftId);
            if (ctrl != msg.sender) revert NotINftController(iNftId, msg.sender);
        }

        IERC20(address(usdc)).safeTransferFrom(msg.sender, address(this), amount);

        verifiers[msg.sender] = Verifier({
            stake: amount,
            votesTotal: 0,
            votesCorrect: 0,
            iNftId: iNftId,
            withdrawableAt: 0,
            active: true
        });
        isRegistered[msg.sender] = true;
        verifierAddresses.push(msg.sender);

        emit VerifierRegistered(msg.sender, amount, iNftId);
    }

    function addStake(
        uint256 amount
    ) external {
        if (!isRegistered[msg.sender]) revert NotRegistered();
        IERC20(address(usdc)).safeTransferFrom(msg.sender, address(this), amount);
        Verifier storage v = verifiers[msg.sender];
        v.stake += amount;
        // If a slash brought us below minStake before, topping up reactivates.
        if (!v.active && v.stake >= minStake) {
            v.active = true;
        }
        emit StakeAdded(msg.sender, amount, v.stake);
    }

    function requestWithdraw() external {
        if (!isRegistered[msg.sender]) revert NotRegistered();
        Verifier storage v = verifiers[msg.sender];
        v.active = false;
        v.withdrawableAt = block.number + WITHDRAW_COOLDOWN_BLOCKS;
        emit WithdrawalRequested(msg.sender, v.withdrawableAt);
    }

    function withdraw() external {
        if (!isRegistered[msg.sender]) revert NotRegistered();
        Verifier storage v = verifiers[msg.sender];
        if (v.withdrawableAt == 0) revert WithdrawNotRequested();
        if (block.number < v.withdrawableAt) {
            revert CooldownNotElapsed(v.withdrawableAt, block.number);
        }
        uint256 amount = v.stake;
        v.stake = 0;
        v.withdrawableAt = 0;
        v.active = false;
        IERC20(address(usdc)).safeTransfer(msg.sender, amount);
        emit StakeWithdrawn(msg.sender, amount);
    }

    /// @inheritdoc IVerifierRegistry
    function slash(
        address verifier,
        uint256 amount,
        address recipient
    ) external returns (uint256 actualSlashed) {
        if (msg.sender != quorum) revert NotQuorum(msg.sender);
        if (!isRegistered[verifier]) revert NotRegistered();
        if (recipient == address(0)) revert ZeroRecipient();

        Verifier storage v = verifiers[verifier];
        // Bounded by current stake — never underflow.
        actualSlashed = amount > v.stake ? v.stake : amount;
        v.stake -= actualSlashed;

        if (v.stake < minStake) {
            v.active = false;
        }

        // Move the slashed USDC out of the registry to the recipient
        // (treasury for verifier dissent; client/majority verifiers for
        // executor FAIL split.
        if (actualSlashed > 0) {
            IERC20(address(usdc)).safeTransfer(recipient, actualSlashed);
        }

        emit VerifierSlashed(verifier, recipient, actualSlashed, v.stake);
    }

    /// @inheritdoc IVerifierRegistry
    function recordVote(address verifier, bool wasCorrect) external {
        if (msg.sender != quorum) revert NotQuorum(msg.sender);
        if (!isRegistered[verifier]) revert NotRegistered();

        Verifier storage v = verifiers[verifier];
        v.votesTotal += 1;
        if (wasCorrect) {
            v.votesCorrect += 1;
        }
        emit VoteRecorded(verifier, wasCorrect, v.votesTotal, v.votesCorrect);
    }

    /// @inheritdoc IVerifierRegistry
    function isActive(
        address verifier
    ) external view returns (bool) {
        if (!isRegistered[verifier]) return false;
        Verifier storage v = verifiers[verifier];
        return v.active && v.stake >= minStake;
    }

    /// @inheritdoc IVerifierRegistry
    function stakeOf(
        address verifier
    ) external view returns (uint256) {
        return verifiers[verifier].stake;
    }

    function getAccuracy(
        address verifier
    ) external view returns (uint256 bps) {
        Verifier storage v = verifiers[verifier];
        return AegisMath.accuracyBps(v.votesCorrect, v.votesTotal);
    }

    function verifierCount() external view returns (uint256) {
        return verifierAddresses.length;
    }
}
