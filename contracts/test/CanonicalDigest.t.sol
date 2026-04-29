// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import {Test} from "forge-std/Test.sol";

contract CanonicalDigestTest is Test {
    /// @notice Canonical JSON of the test ClaimSpec used by Go's
    /// TestID_KnownVector. Identical bytes produced by both sides:
    ///   action: "uniswap_v3_swap", chain_id: 8453, ...
    bytes internal constant SPEC_CANONICAL =
        bytes(
            '{"action":"uniswap_v3_swap","amount_in":"10000000000","chain_id":8453,'
            '"fee_tier":500,"max_slippage_bps":30,"reference_quote_amount_out":"2848000000000000000",'
            '"reference_quote_block":11999000,"token_in":"0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",'
            '"token_out":"0x4200000000000000000000000000000000000006"}'
        );

    function test_KeccakOfCanonicalSpec_MatchesGo() public {
        bytes32 got = keccak256(SPEC_CANONICAL);
        bytes32 expected = 0xab6fc7716fed764bea6d8b7b53569c92f32e75d41d58a35450009bdafc428f36;
        assertEq(
            got,
            expected,
            "canonical JSON keccak diverged - re-anchor BOTH sides"
        );
    }

    /// @notice Locks the recipe for a full ClaimID (canonical || "|" ||
    /// executor || "|" || nonce).
    function test_FullClaimID_MatchesGoRecipe() public pure {
        address executor = address(0x000000000000000000000000000000000000bEEF);
        bytes16 nonce = bytes16(0);

        bytes memory packed = abi.encodePacked(
            SPEC_CANONICAL,
            "|",
            executor,
            "|",
            nonce
        );
        bytes32 got = keccak256(packed);

        // Anchored to Go's TestID_KnownVector output for the same inputs.
        bytes32 expected = 0x6679ed880ab2a9070fe0f2f29187daaba4610b818ecfc71bddae2ecf9f911b9b;
        assertEq(
            got,
            expected,
            "claim ID recipe diverged from Go side - re-anchor BOTH"
        );
    }
}
