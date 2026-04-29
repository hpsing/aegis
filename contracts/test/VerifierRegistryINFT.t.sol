// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Test } from "forge-std/Test.sol";

import { VerifierRegistry } from "../src/VerifierRegistry.sol";
import { VerifierINFT } from "../src/VerifierINFT.sol";
import { IUSDC } from "../src/interfaces/IUSDC.sol";
import { MockUSDC } from "./mocks/MockUSDC.sol";

/// @notice Tests the iNFT-aware registration path added in step 6.
contract VerifierRegistryINFTTest is Test {
    MockUSDC internal usdc;
    VerifierRegistry internal registry;
    VerifierINFT internal inft;

    address internal owner = address(0xA1);
    address internal alice = address(0xA11CE);
    address internal bob = address(0xB0B);
    address internal pretendAegis = address(0xCAFE);

    function setUp() public {
        usdc = new MockUSDC();
        vm.startPrank(owner);
        registry = new VerifierRegistry(IUSDC(address(usdc)), owner);
        inft = new VerifierINFT(owner);
        registry.setAegis(pretendAegis);
        registry.setINftContract(inft);
        vm.stopPrank();

        usdc.mint(alice, 10_000e6);
        usdc.mint(bob, 10_000e6);
    }

    function test_Register_RequiresINftController() public {
        // Alice mints an iNFT.
        vm.prank(alice);
        uint256 tokenId = inft.mintVerifier(alice, "ipfs://card-alice", "noenc");

        // Bob tries to register claiming Alice's iNFT — must revert.
        vm.startPrank(bob);
        usdc.approve(address(registry), 200e6);
        vm.expectRevert(
            abi.encodeWithSelector(VerifierRegistry.NotINftController.selector, tokenId, bob)
        );
        registry.register(200e6, tokenId);
        vm.stopPrank();

        // Alice registers with HER iNFT — succeeds.
        vm.startPrank(alice);
        usdc.approve(address(registry), 200e6);
        registry.register(200e6, tokenId);
        vm.stopPrank();
        assertTrue(registry.isActive(alice));
    }

    function test_Register_INftIdZeroSkipsCheck() public {
        // Alice registers with iNftId=0 (no iNFT linked) — succeeds even
        // without minting first.
        vm.startPrank(alice);
        usdc.approve(address(registry), 200e6);
        registry.register(200e6, 0);
        vm.stopPrank();
        assertTrue(registry.isActive(alice));
    }

    function test_Register_INftCheckSkipped_WhenContractUnset() public {
        // Spin up a fresh registry without setINftContract.
        vm.startPrank(owner);
        VerifierRegistry r2 = new VerifierRegistry(IUSDC(address(usdc)), owner);
        r2.setAegis(pretendAegis);
        vm.stopPrank();

        // Alice registers with arbitrary tokenId — accepted, no check.
        vm.startPrank(alice);
        usdc.approve(address(r2), 200e6);
        r2.register(200e6, 99_999);
        vm.stopPrank();
        assertTrue(r2.isActive(alice));
    }
}
