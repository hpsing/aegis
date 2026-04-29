// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Script, console2 } from "forge-std/Script.sol";

import { AegisContract } from "../src/AegisContract.sol";
import { VerifierRegistry } from "../src/VerifierRegistry.sol";
import { VerifierINFT } from "../src/VerifierINFT.sol";
import { IUSDC } from "../src/interfaces/IUSDC.sol";
import { MockUSDC } from "../test/mocks/MockUSDC.sol";

/// @notice Deploys the full Aegis stack to a public testnet (defaults
/// configured for 0G Galileo). Reads RPC + key from env.
///
/// Run (after `make abigen`):
///
///   export OG_GALILEO_RPC=https://...
///   export DEPLOYER_PRIVATE_KEY=0x...
///   export TREASURY_ADDRESS=0x...
///   export USDC_ADDRESS=0x...   # optional; if unset, deploys MockUSDC
///   forge script script/DeployToOG.s.sol:DeployToOG \
///       --broadcast --rpc-url $OG_GALILEO_RPC \
///       --private-key $DEPLOYER_PRIVATE_KEY
///
/// Outputs deployments/og-galileo.json with all four addresses + the
/// chain id we deployed against. Step 9 picks this up to wire the demo's
/// runtime config on the VPSes.
contract DeployToOG is Script {
    function run() external {
        address treasury = vm.envAddress("TREASURY_ADDRESS");
        address deployer = vm.addr(vm.envUint("DEPLOYER_PRIVATE_KEY"));

        // USDC: prefer the canonical token if the env points at one, else
        // mint a MockUSDC so the demo can mint freely on testnet.
        address usdcAddr;
        try vm.envAddress("USDC_ADDRESS") returns (address u) {
            usdcAddr = u;
        } catch {
            usdcAddr = address(0);
        }

        vm.startBroadcast();

        if (usdcAddr == address(0)) {
            MockUSDC mock = new MockUSDC();
            usdcAddr = address(mock);
            console2.log("Deployed MockUSDC at", usdcAddr);
        } else {
            console2.log("Using existing USDC at", usdcAddr);
        }

        VerifierINFT inft = new VerifierINFT(deployer);
        VerifierRegistry registry = new VerifierRegistry(IUSDC(usdcAddr), deployer);
        AegisContract aegis = new AegisContract(IUSDC(usdcAddr), registry, treasury, deployer);
        registry.setAegis(address(aegis));
        registry.setINftContract(inft);

        vm.stopBroadcast();

        console2.log("USDC:             ", usdcAddr);
        console2.log("VerifierINFT:     ", address(inft));
        console2.log("VerifierRegistry: ", address(registry));
        console2.log("AegisContract:   ", address(aegis));
        console2.log("Treasury:         ", treasury);

        _writeDeployments(usdcAddr, address(inft), address(registry), address(aegis), treasury);
    }

    /// @dev Split out of run() to keep the stack shallow (Solidity hits
    /// "stack too deep" otherwise even with the optimizer on).
    function _writeDeployments(
        address usdcAddr,
        address inft,
        address registry,
        address aegis,
        address treasury
    ) internal {
        string memory part1 = string(
            abi.encodePacked(
                "{\n",
                '  "chain": "',
                vm.toString(block.chainid),
                '",\n',
                '  "usdc": "',
                vm.toString(usdcAddr),
                '",\n',
                '  "inft": "',
                vm.toString(inft),
                '",\n'
            )
        );
        string memory part2 = string(
            abi.encodePacked(
                '  "registry": "',
                vm.toString(registry),
                '",\n',
                '  "aegis": "',
                vm.toString(aegis),
                '",\n',
                '  "treasury": "',
                vm.toString(treasury),
                '"\n',
                "}\n"
            )
        );
        vm.writeFile("deployments/og-galileo.json", string(abi.encodePacked(part1, part2)));
    }
}
