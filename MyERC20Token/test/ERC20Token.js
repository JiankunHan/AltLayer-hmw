const ERC20Token = artifacts.require("ERC20Token");

contract("ERC20Token", (accounts) => {
  it("should deploy with the correct initial supply", async () => {
    let token = await ERC20Token.deployed();
    let balance = await token.balanceOf(accounts[0]);
    assert.equal(balance.toString(), "1000000000000000000000000", "Initial supply is incorrect");
  });
});
