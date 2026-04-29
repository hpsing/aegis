// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Test } from "forge-std/Test.sol";

import { VerifierRegistry } from "../src/VerifierRegistry.sol";
import { IUSDC } from "../src/interfaces/IUSDC.sol";
import { MockUSDC } from "./mocks/MockUSDC.sol";

/// @notice Revert-path coverage for VerifierRegistry.
contract VerifierRegistryEdges is Test {
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
        vm.startPrank(alice);
        usdc.approve(address(registry), type(uint256).max);
        vm.stopPrank();
    }

    function test_AlreadyRegisteredReverts() public {
        vm.prank(alice);
        registry.register(500e6, 0);
        vm.prank(alice);
        vm.expectRevert(VerifierRegistry.AlreadyRegistered.selector);
        registry.register(500e6, 0);
    }

    function test_AddStakeRequiresRegistration() public {
        vm.prank(alice);
        vm.expectRevert(VerifierRegistry.NotRegistered.selector);
        registry.addStake(50e6);
    }

    function test_RequestWithdraw_RequiresRegistration() public {
        vm.prank(alice);
        vm.expectRevert(VerifierRegistry.NotRegistered.selector);
        registry.requestWithdraw();
    }

    function test_Withdraw_RequiresRequest() public {
        vm.prank(alice);
        registry.register(500e6, 0);
        // Direct withdraw without requestWithdraw.
        vm.prank(alice);
        vm.expectRevert(VerifierRegistry.WithdrawNotRequested.selector);
        registry.withdraw();
    }

    function test_AddStake_ReactivatesAfterSlashedOut() public {
        vm.prank(alice);
        registry.register(120e6, 0);
        // Slash to under minStake.
        vm.prank(pretendAegis);
        registry.slash(alice, 30e6, address(0xDEAD));
        assertFalse(registry.isActive(alice), "should be inactive after sub-min slash");

        // Top up; should reactivate.
        vm.prank(alice);
        registry.addStake(20e6);
        assertTrue(registry.isActive(alice), "addStake should reactivate when above min");
    }

    function test_SetMinStake_OnlyOwner() public {
        vm.prank(alice);
        vm.expectRevert();
        registry.setMinStake(50e6);
        vm.prank(owner);
        registry.setMinStake(50e6);
        assertEq(registry.minStake(), 50e6);
    }

    function test_SetAegis_OnlyOnce() public {
        // Already set in setUp(); a second call should revert.
        vm.prank(owner);
        vm.expectRevert(bytes("aegis already set"));
        registry.setAegis(address(0xBEEF));
    }

    function test_SetAegis_RejectsZero() public {
        // Fresh registry where aegis hasn't been set.
        vm.prank(owner);
        VerifierRegistry r = new VerifierRegistry(IUSDC(address(usdc)), owner);
        vm.prank(owner);
        vm.expectRevert(bytes("zero aegis"));
        r.setAegis(address(0));
    }

    function test_VerifierCount_GrowsOnRegister() public {
        assertEq(registry.verifierCount(), 0);
        vm.prank(alice);
        registry.register(500e6, 0);
        assertEq(registry.verifierCount(), 1);
    }

    function test_GetAccuracy_ZeroVotesIsZero() public {
        vm.prank(alice);
        registry.register(500e6, 0);
        assertEq(registry.getAccuracy(alice), 0);
    }

    function test_Slash_OnUnregisteredReverts() public {
        vm.prank(pretendAegis);
        vm.expectRevert(VerifierRegistry.NotRegistered.selector);
        registry.slash(alice, 10e6, address(0xDEAD));
    }

    function test_Slash_ZeroRecipientReverts() public {
        vm.prank(alice);
        registry.register(500e6, 0);
        vm.prank(pretendAegis);
        vm.expectRevert(VerifierRegistry.ZeroRecipient.selector);
        registry.slash(alice, 10e6, address(0));
    }

    function test_RecordVote_OnUnregisteredReverts() public {
        vm.prank(pretendAegis);
        vm.expectRevert(VerifierRegistry.NotRegistered.selector);
        registry.recordVote(alice, true);
    }
}
