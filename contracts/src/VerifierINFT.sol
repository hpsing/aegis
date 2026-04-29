// SPDX-License-Identifier: MIT
pragma solidity 0.8.24;

import { ERC721 } from "@openzeppelin/contracts/token/ERC721/ERC721.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

/// @title VerifierINFT — ERC-7857 verifier identity (minimal impl)
/// @notice Each verifier mints one iNFT during registration. The token
/// carries:
///   - agentCardURI            — JSON metadata on 0G Storage describing
///                                the verifier's capabilities
///   - encryptedMetadataURI    — pointer to encrypted weights / personality
///                                (TODO:: later)
///   - controller              — mutable EOA that operates the verifier;
///                                the iNFT owner can change this when the
///                                token is sold/inherited (e.g. on AIverse)
contract VerifierINFT is ERC721, Ownable {
    uint256 private _nextTokenId = 1;

    /// @notice Per-token metadata. agentCardURI and encryptedMetadataURI
    /// can be updated by the current controller (not the holder) — this
    /// lets a transferred token's new operator publish their own card
    /// without needing the previous holder to coordinate.
    struct Meta {
        string agentCardURI;
        string encryptedMetadataURI;
        address controller;
    }

    mapping(uint256 => Meta) private _meta;

    event VerifierMinted(
        uint256 indexed tokenId, address indexed to, string agentCardURI, string encryptedMetadataURI
    );
    event ControllerUpdated(uint256 indexed tokenId, address indexed oldController, address indexed newController);
    event AgentCardUpdated(uint256 indexed tokenId, string newURI);
    event EncryptedMetadataUpdated(uint256 indexed tokenId, string newURI);

    error NotControllerOrHolder();
    error InvalidController();

    constructor(
        address initialOwner
    ) ERC721("Aegis Verifier iNFT", "AVI") Ownable(initialOwner) { }

    /// @notice Mint a new verifier iNFT to `to`. `to` becomes both the
    /// holder AND the initial controller — two roles that diverge later
    /// when the token is sold.
    /// Anyone can mint for themselves; we don't gate this with a
    /// permission since the *registration* check (in VerifierRegistry)
    /// is what actually grants the iNFT economic privileges.
    function mintVerifier(
        address to,
        string calldata agentCardURI,
        string calldata encryptedMetadataURI
    ) external returns (uint256 tokenId) {
        if (to == address(0)) revert InvalidController();
        tokenId = _nextTokenId++;
        _safeMint(to, tokenId);
        _meta[tokenId] = Meta({
            agentCardURI: agentCardURI,
            encryptedMetadataURI: encryptedMetadataURI,
            controller: to
        });
        emit VerifierMinted(tokenId, to, agentCardURI, encryptedMetadataURI);
    }

    /// @notice Set the operator EOA. Callable by the current holder (after
    /// transfer of the token) — i.e. the typical flow is:
    ///   1. seller transfers the iNFT to the buyer
    ///   2. buyer calls setController to point the iNFT at THEIR EOA
    ///   3. buyer's EOA can now register as a verifier with this iNFT
    function setController(uint256 tokenId, address newController) external {
        if (newController == address(0)) revert InvalidController();
        if (msg.sender != ownerOf(tokenId)) revert NotControllerOrHolder();
        address old = _meta[tokenId].controller;
        _meta[tokenId].controller = newController;
        emit ControllerUpdated(tokenId, old, newController);
    }

    /// @notice Update the agent card URI. The CONTROLLER (not necessarily
    /// the holder) can publish a new card — this is the operator's
    /// running configuration.
    function setAgentCardURI(uint256 tokenId, string calldata newURI) external {
        if (msg.sender != _meta[tokenId].controller) revert NotControllerOrHolder();
        _meta[tokenId].agentCardURI = newURI;
        emit AgentCardUpdated(tokenId, newURI);
    }

    function setEncryptedMetadataURI(uint256 tokenId, string calldata newURI) external {
        if (msg.sender != _meta[tokenId].controller) revert NotControllerOrHolder();
        _meta[tokenId].encryptedMetadataURI = newURI;
        emit EncryptedMetadataUpdated(tokenId, newURI);
    }

    function controllerOf(
        uint256 tokenId
    ) external view returns (address) {
        return _meta[tokenId].controller;
    }

    function agentCardURI(
        uint256 tokenId
    ) external view returns (string memory) {
        return _meta[tokenId].agentCardURI;
    }

    function encryptedMetadataURI(
        uint256 tokenId
    ) external view returns (string memory) {
        return _meta[tokenId].encryptedMetadataURI;
    }
}
