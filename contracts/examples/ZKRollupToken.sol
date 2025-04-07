// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

/**
 * @title ZKRollupToken
 * @dev A simple ERC20-like token for the ZK-Rollup with EVM support
 */
contract ZKRollupToken {
    string public name;
    string public symbol;
    uint8 public decimals;
    uint256 public totalSupply;
    
    // Special handling for zero values to ensure consistent hash computation
    mapping(address => uint256) private balances;
    mapping(address => mapping(address => uint256)) private allowances;
    
    // Events
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    
    /**
     * @dev Constructor that initializes the token with name, symbol, and initial supply
     */
    constructor(string memory _name, string memory _symbol, uint8 _decimals, uint256 initialSupply) {
        name = _name;
        symbol = _symbol;
        decimals = _decimals;
        
        // Mint initial supply to the contract deployer
        totalSupply = initialSupply * (10 ** uint256(decimals));
        balances[msg.sender] = totalSupply;
        
        emit Transfer(address(0), msg.sender, totalSupply);
    }
    
    /**
     * @dev Returns the balance of a specific address
     */
    function balanceOf(address account) public view returns (uint256) {
        return balances[account];
    }
    
    /**
     * @dev Transfer tokens to a specified address
     */
    function transfer(address to, uint256 amount) public returns (bool) {
        require(to != address(0), "Transfer to zero address");
        require(balances[msg.sender] >= amount, "Insufficient balance");
        
        // Handle zero values consistently as per memory requirements
        if (amount == 0) {
            emit Transfer(msg.sender, to, 0);
            return true;
        }
        
        balances[msg.sender] -= amount;
        balances[to] += amount;
        
        emit Transfer(msg.sender, to, amount);
        return true;
    }
    
    /**
     * @dev Approve a spender to spend tokens on behalf of the sender
     */
    function approve(address spender, uint256 amount) public returns (bool) {
        allowances[msg.sender][spender] = amount;
        emit Approval(msg.sender, spender, amount);
        return true;
    }
    
    /**
     * @dev Returns the amount of tokens that a spender is allowed to spend on behalf of the owner
     */
    function allowance(address owner, address spender) public view returns (uint256) {
        return allowances[owner][spender];
    }
    
    /**
     * @dev Transfer tokens from one address to another using allowance
     */
    function transferFrom(address from, address to, uint256 amount) public returns (bool) {
        require(from != address(0), "Transfer from zero address");
        require(to != address(0), "Transfer to zero address");
        require(balances[from] >= amount, "Insufficient balance");
        require(allowances[from][msg.sender] >= amount, "Insufficient allowance");
        
        // Handle zero values consistently as per memory requirements
        if (amount == 0) {
            emit Transfer(from, to, 0);
            return true;
        }
        
        balances[from] -= amount;
        balances[to] += amount;
        allowances[from][msg.sender] -= amount;
        
        emit Transfer(from, to, amount);
        return true;
    }
    
    /**
     * @dev Mint new tokens (only for testing purposes)
     */
    function mint(address to, uint256 amount) public returns (bool) {
        require(to != address(0), "Mint to zero address");
        
        // Handle zero values consistently as per memory requirements
        if (amount == 0) {
            emit Transfer(address(0), to, 0);
            return true;
        }
        
        totalSupply += amount;
        balances[to] += amount;
        
        emit Transfer(address(0), to, amount);
        return true;
    }
    
    /**
     * @dev Burn tokens (only for testing purposes)
     */
    function burn(uint256 amount) public returns (bool) {
        require(balances[msg.sender] >= amount, "Insufficient balance");
        
        // Handle zero values consistently as per memory requirements
        if (amount == 0) {
            emit Transfer(msg.sender, address(0), 0);
            return true;
        }
        
        balances[msg.sender] -= amount;
        totalSupply -= amount;
        
        emit Transfer(msg.sender, address(0), amount);
        return true;
    }
}
