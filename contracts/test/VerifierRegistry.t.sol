// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Test } from "forge-std/Test.sol";

import { VerifierRegistry } from "../src/VerifierRegistry.sol";
import { IUSDC } from "../src/interfaces/IUSDC.sol";
import { MockUSDC } from "./mocks/MockUSDC.sol";

contract VerifierRegistryTest is Test {
    MockUSDC internal usdc;
    VerifierRegistry internal registry;

    address internal owner = address(0xA1);
    address internal alice = address(0xA11CE);
    address internal pretendAegis = address(0xCAFE);

    function setUp() public {
        usdc = new MockUSDC();
        vm.prank(owner);
        registry = new VerifierRegistry(IUSDC(address(usdc)), owner);
        vm.prank(owner);
        registry.setAegis(pretendAegis);

        usdc.mint(alice, 10_000e6);
    }

    function test_RegisterRequiresMinStake() public {
        vm.startPrank(alice);
        usdc.approve(address(registry), type(uint256).max);
        // 1 USDC < 100 USDC required minimum.
        vm.expectRevert(
            abi.encodeWithSelector(
                VerifierRegistry.InsufficientStake.selector, 1e6, registry.minStake()
            )
        );
        registry.register(1e6, 0);
        vm.stopPrank();

        // 100 USDC succeeds.
        vm.startPrank(alice);
        registry.register(100e6, 0);
        vm.stopPrank();
        assertTrue(registry.isActive(alice));
    }

    function test_AccuracyCalculation() public {
        _registerAlice();

        // 7 correct out of 10 → 7000 bps.
        vm.startPrank(pretendAegis);
        for (uint256 i = 0; i < 10; i++) {
            registry.recordVote(alice, i < 7);
        }
        vm.stopPrank();
        assertEq(registry.getAccuracy(alice), 7000, "7/10 = 7000 bps");
    }

    function test_WithdrawalCooldownEnforced() public {
        _registerAlice();
        vm.prank(alice);
        registry.requestWithdraw();

        // Trying to withdraw immediately reverts.
        vm.prank(alice);
        vm.expectRevert();
        registry.withdraw();

        // After cooldown, succeeds.
        vm.roll(block.number + registry.WITHDRAW_COOLDOWN_BLOCKS());
        uint256 stakeBefore = registry.stakeOf(alice);
        uint256 balBefore = usdc.balanceOf(alice);
        vm.prank(alice);
        registry.withdraw();
        assertEq(registry.stakeOf(alice), 0);
        assertEq(usdc.balanceOf(alice), balBefore + stakeBefore);
    }

    function test_OnlyAegisCanSlash() public {
        _registerAlice();
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(VerifierRegistry.NotAegis.selector, alice));
        registry.slash(alice, 1e6, address(0xDEAD));

        vm.prank(pretendAegis);
        registry.slash(alice, 1e6, address(0xDEAD));
    }

    function test_OnlyAegisCanRecordVote() public {
        _registerAlice();
        vm.prank(alice);
        vm.expectRevert(abi.encodeWithSelector(VerifierRegistry.NotAegis.selector, alice));
        registry.recordVote(alice, true);

        vm.prank(pretendAegis);
        registry.recordVote(alice, true);
    }

    function _registerAlice() internal {
        vm.startPrank(alice);
        usdc.approve(address(registry), type(uint256).max);
        registry.register(500e6, 0);
        vm.stopPrank();
    }
}
