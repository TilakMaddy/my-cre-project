// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

contract Receiver {
    bytes32 public immutable i_workflowID;
    address public immutable i_forwarder;

    mapping(bytes32 => bool) public processedNonces;

    event Settled(address indexed recipient, uint256 amount, bytes32 nonce);

    error InvalidSender();
    error AlreadyProcessed();
    error StaleReport();

    constructor(bytes32 workflowID, address forwarder) {
        i_workflowID = workflowID;
        i_forwarder = forwarder;
    }

    struct Report {
        bytes32 workflowId;
        uint64  timestamp;
        address recipient;
        uint256 amount;
        bytes32 nonce;
    }

    function onReport(bytes calldata report) external {
        if (msg.sender != i_forwarder) revert InvalidSender();

        Report memory r = abi.decode(report, (Report));

        if (r.workflowId != i_workflowID) revert InvalidSender();
        if (processedNonces[r.nonce]) revert AlreadyProcessed();
        if (r.timestamp + 3600 < block.timestamp) revert StaleReport();

        processedNonces[r.nonce] = true;
        emit Settled(r.recipient, r.amount, r.nonce);
    }
}
