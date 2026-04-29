// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

interface IReentrancyTarget {
    function settle(
        uint256 jobId
    ) external;
}

/// @notice A USDC stand-in that, on transfer to a configured trigger
/// address, calls back into AegisContract.settle(jobId). Used to prove
/// AegisContract's nonReentrant guard actually fires.
contract ReentrantUSDC is ERC20 {
    address public reentrancyTrigger;
    address public aegis;
    uint256 public reentryJobId;
    bool public attemptedReenter;
    bool public reenterReverted;

    constructor() ERC20("ReentrantUSDC", "rUSDC") { }

    function decimals() public pure override returns (uint8) {
        return 6;
    }

    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    /// @notice Configure the recipient address that triggers a reenter,
    /// and the AegisContract + jobId to call back into.
    function arm(address trigger, address _aegis, uint256 _jobId) external {
        reentrancyTrigger = trigger;
        aegis = _aegis;
        reentryJobId = _jobId;
    }

    function _update(address from, address to, uint256 amount) internal override {
        super._update(from, to, amount);
        if (to != address(0) && to == reentrancyTrigger && aegis != address(0)) {
            attemptedReenter = true;
            // Prevent infinite recursion in case the guard is missing.
            address localAegis = aegis;
            aegis = address(0);
            try IReentrancyTarget(localAegis).settle(reentryJobId) {
                reenterReverted = false;
            } catch {
                reenterReverted = true;
            }
        }
    }
}
