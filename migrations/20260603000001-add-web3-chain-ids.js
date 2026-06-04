/**
 * Sets `config.allowed_web3_chain_ids` to an empty array on all existing
 * namespaces. Web3 sign-in is opt-in: an admin must populate the array
 * (or rely on the global default at sign-in time) before wallet logins work.
 */
module.exports = {
  async up(db) {
    await db.collection("namespaces").updateMany(
      { "config.allowed_web3_chain_ids": { $exists: false } },
      { $set: { "config.allowed_web3_chain_ids": [] } }
    );
  },

  async down(db) {
    await db.collection("namespaces").updateMany(
      {},
      { $unset: { "config.allowed_web3_chain_ids": "" } }
    );
  },
};
