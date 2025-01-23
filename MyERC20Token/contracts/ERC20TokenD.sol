// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

// 引入 OpenZeppelin 的 ERC20 合约实现
import "../node_modules/@openzeppelin/contracts/token/ERC20/ERC20.sol";

contract ERC20TokenD is ERC20 {
    constructor(uint256 initialSupply) ERC20("MyToken", "MTK") {
        _mint(msg.sender, initialSupply);  // 为创建者铸造初始供应量
    }
}
