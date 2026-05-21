/**
 * Migrates the namespace config from boolean `mfa_required` to string `mfa_policy`.
 *
 * Mapping:
 *   mfa_required: true  → mfa_policy: "required"
 *   mfa_required: false → mfa_policy: "disabled"
 *
 * The old `mfa_required` field is removed after migration.
 */
module.exports = {
  async up(db) {
    // Set mfa_policy = "required" where mfa_required was true
    await db.collection("namespaces").updateMany(
      { "config.mfa_required": true },
      {
        $set: { "config.mfa_policy": "required" },
        $unset: { "config.mfa_required": "" },
      }
    );

    // Set mfa_policy = "disabled" where mfa_required was false or missing
    await db.collection("namespaces").updateMany(
      { "config.mfa_policy": { $exists: false } },
      {
        $set: { "config.mfa_policy": "disabled" },
        $unset: { "config.mfa_required": "" },
      }
    );
  },

  async down(db) {
    // Reverse: map mfa_policy back to mfa_required boolean
    await db.collection("namespaces").updateMany(
      { "config.mfa_policy": "required" },
      {
        $set: { "config.mfa_required": true },
        $unset: { "config.mfa_policy": "" },
      }
    );

    await db.collection("namespaces").updateMany(
      { "config.mfa_policy": { $in: ["disabled", "optional"] } },
      {
        $set: { "config.mfa_required": false },
        $unset: { "config.mfa_policy": "" },
      }
    );

    // Catch any remaining documents
    await db.collection("namespaces").updateMany(
      { "config.mfa_required": { $exists: false } },
      { $set: { "config.mfa_required": false } }
    );
  },
};
