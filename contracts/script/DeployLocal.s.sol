// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { Script, console2 } from "forge-std/Script.sol";

import { AegisContract } from "../src/AegisContract.sol";
import { VerifierRegistry } from "../src/VerifierRegistry.sol";
import { VerifierINFT } from "../src/VerifierINFT.sol";
import { IUSDC } from "../src/interfaces/IUSDC.sol";
import { MockUSDC } from "../test/mocks/MockUSDC.sol";

/// @notice Local Anvil deployment. Mints + funds Anvil's first 4
/// pre-funded accounts so they can play client/executor/verifier roles
/// in step 5's integration tests.
///
/// Run:
///   anvil &
///   forge script script/DeployLocal.s.sol:DeployLocal --broadcast \
///       --rpc-url http://127.0.0.1:8545 \
///       --private-key 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
contract DeployLocal is Script {
    /// @notice Anvil's first 5 default accounts, pre-funded with 10K ETH each.
    address internal constant DEPLOYER = 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266;
    address internal constant CLIENT = 0x70997970C51812dc3A010C7d01b50e0d17dc79C8;
    address internal constant EXECUTOR = 0x3C44CdDdB6a900fa2b585dd299e03d12FA4293BC;
    address internal constant VERIFIER1 = 0x90F79bf6EB2c4f870365E785982E1f101E93b906;
    address internal constant TREASURY = 0x976EA74026E726554dB657fA54763abd0C3a0aa9; // Anvil #6 (avoids collision with verifier slots)

    function run() external {
        vm.startBroadcast();

        MockUSDC usdc = new MockUSDC();
        VerifierINFT inft = new VerifierINFT(DEPLOYER);
        VerifierRegistry registry = new VerifierRegistry(IUSDC(address(usdc)), DEPLOYER);
        AegisContract aegis =
            new AegisContract(IUSDC(address(usdc)), registry, TREASURY, DEPLOYER);
        registry.setAegis(address(aegis));
        registry.setINftContract(inft);

        // Fund the demo actors so they can post jobs / register / etc.
        usdc.mint(DEPLOYER, 1_000_000e6);
        usdc.mint(CLIENT, 100_000e6);
        usdc.mint(EXECUTOR, 100_000e6);
        usdc.mint(VERIFIER1, 100_000e6);

        vm.stopBroadcast();

        console2.log("MockUSDC:        ", address(usdc));
        console2.log("VerifierINFT:    ", address(inft));
        console2.log("VerifierRegistry:", address(registry));
        console2.log("AegisContract:  ", address(aegis));
        console2.log("Treasury:        ", TREASURY);

        string memory json = string(
            abi.encodePacked(
                "{\n",
                '  "chain": "anvil-31337",\n',
                '  "usdc": "',
                vm.toString(address(usdc)),
                '",\n',
                '  "inft": "',
                vm.toString(address(inft)),
                '",\n',
                '  "registry": "',
                vm.toString(address(registry)),
                '",\n',
                '  "aegis": "',
                vm.toString(address(aegis)),
                '"\n',
                "}\n"
            )
        );
        vm.writeFile("deployments/local.json", json);
    }
}
