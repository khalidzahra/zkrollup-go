// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract SimpleStorage {
    uint256 private value;
    
    // Event to emit when the value is updated
    event ValueChanged(uint256 newValue);
    
    // Constructor sets the initial value
    constructor(uint256 initialValue) {
        value = initialValue;
    }
    
    // Set a new value
    function setValue(uint256 newValue) public {
        value = newValue;
        emit ValueChanged(newValue);
    }
    
    // Get the current value
    function getValue() public view returns (uint256) {
        return value;
    }
    
    // Increment the current value by a specified amount
    function incrementValue(uint256 amount) public {
        value += amount;
        emit ValueChanged(value);
    }
}
