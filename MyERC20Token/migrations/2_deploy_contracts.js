// const ERC20TokenD = artifacts.require("ERC20TokenD");

// module.exports = async function (deployer) {
//   // 部署合约时传递初始供应量
//   deployer.deploy(ERC20TokenD, web3.utils.toWei('1000000', 'ether')); // 1 million 代币，单位为最小单位（例如 wei）
// };

const WithdrawToken = artifacts.require("WithdrawToken");

module.exports = function (deployer) {
  deployer.deploy(WithdrawToken);
};
