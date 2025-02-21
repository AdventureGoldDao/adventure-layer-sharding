// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract PrecompileCaller {
    function GetTimestampHD() public view returns (uint256) {
        address precompileAddress = address(0x64); // Address of the precompiled contract
        bytes4 targetFuncSelector = bytes4(keccak256("getTimestampHD()")); // Function selector

        // Prepare input data: only the function selector since there are no input parameters
        bytes memory input = abi.encodePacked(targetFuncSelector);
        bytes memory output = new bytes(32); // The return value is uint256, occupying 32 bytes
        bool success;

        assembly {
            // Call the precompiled contract
            success := staticcall(
                gas(),                // Use remaining gas
                precompileAddress,    // Precompiled contract address
                add(input, 0x20),     // Location of input data
                mload(input),         // Length of input data
                add(output, 0x20),    // Location of output data
                32                    // Length of output data (uint256 is 32 bytes)
            )
        }

        require(success, "Precompile call failed");

        // Decode the return value and return uint256 type
        return abi.decode(output, (uint256));
    }
}

