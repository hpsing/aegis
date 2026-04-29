// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Test } from "forge-std/Test.sol";

import { VerifierINFT } from "../src/VerifierINFT.sol";

contract VerifierINFTTest is Test {
    VerifierINFT internal inft;

    address internal owner = address(0xA1);
    address internal alice = address(0xA11CE);
    address internal bob = address(0xB0B);

    function setUp() public {
        vm.prank(owner);
        inft = new VerifierINFT(owner);
    }

    function test_Mint_HolderIsController() public {
        vm.prank(alice);
        uint256 tokenId = inft.mintVerifier(alice, "ipfs://card-1", "placeholder://noenc");

        assertEq(inft.ownerOf(tokenId), alice);
        assertEq(inft.controllerOf(tokenId), alice);
        assertEq(inft.agentCardURI(tokenId), "ipfs://card-1");
    }

    function test_SetController_OnlyHolder() public {
        vm.prank(alice);
        uint256 tokenId = inft.mintVerifier(alice, "card", "noenc");

        // Bob can't change controller of Alice's token.
        vm.prank(bob);
        vm.expectRevert(VerifierINFT.NotControllerOrHolder.selector);
        inft.setController(tokenId, bob);

        // Alice can.
        vm.prank(alice);
        inft.setController(tokenId, bob);
        assertEq(inft.controllerOf(tokenId), bob);
    }

    function test_TransferThenSetController() public {
        vm.prank(alice);
        uint256 tokenId = inft.mintVerifier(alice, "card", "noenc");

        // Alice sells the token to Bob.
        vm.prank(alice);
        inft.transferFrom(alice, bob, tokenId);
        assertEq(inft.ownerOf(tokenId), bob);
        // Controller didn't auto-update — Bob has to call setController.
        assertEq(inft.controllerOf(tokenId), alice);

        // Now Bob (the new holder) updates controller to themselves.
        vm.prank(bob);
        inft.setController(tokenId, bob);
        assertEq(inft.controllerOf(tokenId), bob);
    }

    function test_SetAgentCard_OnlyController() public {
        vm.prank(alice);
        uint256 tokenId = inft.mintVerifier(alice, "card-1", "noenc");

        // Bob can't update the card.
        vm.prank(bob);
        vm.expectRevert(VerifierINFT.NotControllerOrHolder.selector);
        inft.setAgentCardURI(tokenId, "card-mallet");

        // Alice (the controller) can.
        vm.prank(alice);
        inft.setAgentCardURI(tokenId, "card-2");
        assertEq(inft.agentCardURI(tokenId), "card-2");
    }

    function test_RejectsZeroController() public {
        vm.prank(alice);
        uint256 tokenId = inft.mintVerifier(alice, "card", "noenc");

        vm.prank(alice);
        vm.expectRevert(VerifierINFT.InvalidController.selector);
        inft.setController(tokenId, address(0));
    }

    function test_MintRejectsZeroRecipient() public {
        vm.expectRevert(VerifierINFT.InvalidController.selector);
        inft.mintVerifier(address(0), "card", "noenc");
    }
}
